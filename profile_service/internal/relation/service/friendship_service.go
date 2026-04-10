package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"profile_service/internal/kafka"
	"profile_service/internal/relation/models"
	"profile_service/internal/relation/repository"
	"profile_service/internal/relation/service/interfaces"
	"profile_service/internal/user/service"
	"profile_service/internal/user/service/service_dto"
)

type FriendshipService struct {
	friendshipRepo repository.FriendshipRepositoryInterface
	txManager      repository.TransactionManager
	userService    service.UserServiceInterface
	kafkaProducer  kafka.ProducerInterface
	interfaces.UserRelationCheckerInterface
	log *logrus.Entry
}

func NewFriendshipService(friendshipRepo repository.FriendshipRepositoryInterface,
	userService service.UserServiceInterface, txManager repository.TransactionManager,
	kafkaProducer kafka.ProducerInterface, log *logrus.Entry) interfaces.FriendshipServiceInterface {
	return &FriendshipService{
		friendshipRepo: friendshipRepo,
		userService:    userService,
		txManager:      txManager,
		kafkaProducer:  kafkaProducer,
		log:            log,
	}
}

func (f *FriendshipService) SendFriendRequest(ctx context.Context, senderId, receiverId int64, message string) error {
	f.log.WithFields(logrus.Fields{
		"sender_id":   senderId,
		"receiver_id": receiverId,
	}).Info("Sending friend request")
	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		// Проверяем блокировку
		blocked, err := tx.IsBlocked(ctx, senderId, receiverId)
		if err != nil {
			f.log.WithError(err).Error("Failed to check block status")
			return fmt.Errorf("failed to check block status: %w", err)
		}
		if blocked {
			f.log.Warn("User is blocked, cannot send friend request")
			return errors.New("cannot send friend request: user is blocked")
		}

		// Проверяем дружбу
		areFriends, err := tx.AreFriends(ctx, senderId, receiverId)
		if err != nil {
			f.log.WithError(err).Error("Failed to check friendship status")
			return fmt.Errorf("failed to check friendship status: %w", err)
		}
		if areFriends {
			f.log.Warn("Users are already friends")
			return errors.New("users are already friends")
		}

		// Проверяем активный запрос
		existingRequest, err := tx.GetActiveRequestBetweenUsers(ctx, senderId, receiverId)
		if err == nil && existingRequest != nil {
			f.log.Warn("Friend request already exists")
			return errors.New("friend request already exists")
		}

		// Создаем запрос
		request := &models.FriendRequest{
			SenderId:   senderId,
			ReceiverId: receiverId,
			Status:     repository.RequestStatusPending,
			Message:    message,
		}

		if err := tx.CreateFriendRequest(ctx, request); err != nil {
			f.log.WithError(err).Error("Failed to create friend request")
			return fmt.Errorf("failed to create friend request: %w", err)
		}

		// Записываем в историю
		history := &models.FriendshipHistory{
			EventType: "request_sent",
			UserId:    senderId,
			TargetId:  receiverId,
			NewStatus: stringPtr(repository.RequestStatusPending),
			RequestId: &request.Id,
			Metadata:  createMetadata("message", message),
		}

		if err := tx.CreateHistory(ctx, history); err != nil {
			f.log.WithError(err).Warn("Failed to create history entry (non-critical)")
		}

		// Отправляем событие
		event := kafka.NewFriendRequestSentEvent(senderId, receiverId, request.Id, message)
		go func() {
			if err := f.kafkaProducer.SendEvent(context.Background(), "friendship-events",
				fmt.Sprintf("%d", senderId), event); err != nil {
				f.log.WithError(err).Error("Failed to send Kafka event")
			}
		}()

		f.log.WithFields(logrus.Fields{
			"request_id":  request.Id,
			"sender_id":   senderId,
			"receiver_id": receiverId,
		}).Info("Friend request sent successfully")

		return nil

	})
}

func (f *FriendshipService) AnswerFriendRequest(ctx context.Context, requestId, userId int64, accept bool) error {
	f.log.WithFields(logrus.Fields{
		"request_id": requestId,
		"user_id":    userId,
		"accept":     accept,
	}).Info("Answering friend request")
	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		request, err := tx.GetPendingRequest(ctx, requestId, userId)
		if err != nil {
			f.log.WithError(err).Error("Failed to get pending request")
			if errors.Is(err, errors.New("friendship not found")) {
				return errors.New("friend request not found or already processed")
			}
			return fmt.Errorf("failed to get friend request: %w", err)
		}

		if accept {
			if err := tx.UpdateFriendRequestStatus(ctx, requestId, repository.RequestStatusAccepted); err != nil {
				f.log.WithError(err).Error("Failed to update request status")
				return fmt.Errorf("failed to update request status: %w", err)
			}

			friend := &models.Friend{
				UserId:   request.SenderId,
				FriendId: request.ReceiverId,
			}
			if err := tx.CreateFriend(ctx, friend); err != nil {
				f.log.WithError(err).Error("Failed to create friend relationship")
				return fmt.Errorf("failed to create friend relationship: %w", err)
			}

			history := &models.FriendshipHistory{
				EventType: "request_accepted",
				UserId:    userId,
				TargetId:  request.SenderId,
				OldStatus: stringPtr(repository.RequestStatusPending),
				NewStatus: stringPtr(repository.RequestStatusAccepted),
				RequestId: &request.Id,
			}
			if err := tx.CreateHistory(ctx, history); err != nil {
				f.log.WithError(err).Warn("Failed to create history entry (non-critical)")
			}

			// Отправляем событие
			event := kafka.NewFriendRequestActionEvent(userId, request.SenderId, requestId, "accepted")
			go func() {
				if err := f.kafkaProducer.SendEvent(context.Background(), "friendship-events",
					fmt.Sprintf("%d", userId), event); err != nil {
					f.log.WithError(err).Error("Failed to send Kafka event")
				}
			}()

			f.log.WithFields(logrus.Fields{
				"request_id":  requestId,
				"sender_id":   request.SenderId,
				"receiver_id": userId,
			}).Info("Friend request accepted")
		} else {
			if err := tx.UpdateFriendRequestStatus(ctx, requestId, repository.RequestStatusRejected); err != nil {
				f.log.WithError(err).Error("Failed to reject request")
				return fmt.Errorf("failed to reject request: %w", err)
			}

			history := &models.FriendshipHistory{
				EventType: "request_rejected",
				UserId:    userId,
				TargetId:  request.SenderId,
				OldStatus: stringPtr(repository.RequestStatusPending),
				NewStatus: stringPtr(repository.RequestStatusRejected),
				RequestId: &request.Id,
			}
			if err := tx.CreateHistory(ctx, history); err != nil {
				f.log.WithError(err).Warn("Failed to create history entry (non-critical)")
			}

			event := kafka.NewFriendRequestActionEvent(userId, request.SenderId, requestId, "rejected")
			go func() {
				if err := f.kafkaProducer.SendEvent(context.Background(), "friendship-events",
					fmt.Sprintf("%d", userId), event); err != nil {
					f.log.WithError(err).Error("Failed to send Kafka event")
				}
			}()
			f.log.WithFields(logrus.Fields{
				"request_id":  requestId,
				"sender_id":   request.SenderId,
				"receiver_id": userId,
			}).Info("Friend request rejected")
		}
		return nil
	})
}

func (f FriendshipService) BlockUser(ctx context.Context, blockerId, blockedId int64, reason string) error {
	//TODO implement me
	panic("implement me")
}

func (f FriendshipService) UnblockUser(ctx context.Context, blockerId, blockedId int64) error {
	//TODO implement me
	panic("implement me")
}

func (f FriendshipService) DeleteFromFriendList(ctx context.Context, userId, friendId int64) error {
	//TODO implement me
	panic("implement me")
}

func (f FriendshipService) GetFriendList(ctx context.Context, userId int64) (*service_dto.GetUserViewListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (f FriendshipService) CheckRequestState(ctx context.Context, userId, targetId int64) (string, error) {
	//TODO implement me
	panic("implement me")
}

func createMetadata(key, value string) json.RawMessage {
	metadata := map[string]string{key: value}
	data, _ := json.Marshal(metadata)
	return data
}

func stringPtr(s string) *string {
	return &s
}
