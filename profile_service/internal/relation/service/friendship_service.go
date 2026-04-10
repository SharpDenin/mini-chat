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
	friendshipRepo      repository.FriendshipRepositoryInterface
	txManager           repository.TransactionManagerInterface
	userService         service.UserServiceInterface
	outboxKafkaProducer *kafka.OutboxProducer
	log                 *logrus.Entry
}

func NewFriendshipService(friendshipRepo repository.FriendshipRepositoryInterface,
	userService service.UserServiceInterface, txManager repository.TransactionManagerInterface,
	outboxKafkaProducer *kafka.OutboxProducer, log *logrus.Entry) interfaces.FriendshipServiceInterface {
	return &FriendshipService{
		friendshipRepo:      friendshipRepo,
		userService:         userService,
		txManager:           txManager,
		outboxKafkaProducer: outboxKafkaProducer,
		log:                 log,
	}
}

func (f *FriendshipService) SendFriendRequest(ctx context.Context, senderId, receiverId int64, message string) error {
	f.log.WithFields(logrus.Fields{
		"sender_id":   senderId,
		"receiver_id": receiverId,
	}).Info("Sending friend request")
	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		blocked, err := tx.IsBlocked(ctx, senderId, receiverId)
		if err != nil {
			f.log.WithError(err).Error("Failed to check block status")
			return fmt.Errorf("failed to check block status: %w", err)
		}
		if blocked {
			f.log.Warn("User is blocked, cannot send friend request")
			return errors.New("cannot send friend request: user is blocked")
		}

		areFriends, err := tx.AreFriends(ctx, senderId, receiverId)
		if err != nil {
			f.log.WithError(err).Error("Failed to check friendship status")
			return fmt.Errorf("failed to check friendship status: %w", err)
		}
		if areFriends {
			f.log.Warn("Users are already friends")
			return errors.New("users are already friends")
		}

		existingRequest, err := tx.GetActiveRequestBetweenUsers(ctx, senderId, receiverId)
		if err != nil && !errors.Is(err, repository.ErrFriendshipNotFound) {
			f.log.WithError(err).Error("Failed to check existing request")
			return fmt.Errorf("failed to check existing request: %w", err)
		}
		if existingRequest != nil {
			f.log.Warn("Friend request already exists")
			return errors.New("friend request already exists")
		}

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

		event := kafka.NewFriendRequestSentEvent(senderId, receiverId, request.Id, message)
		go func() {
			if err := f.outboxKafkaProducer.SendEvent(context.Background(), "friendship-events",
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

		if request == nil {
			return errors.New("friend request not found")
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
				if err := f.outboxKafkaProducer.SendEvent(context.Background(), "friendship-events",
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
				if err := f.outboxKafkaProducer.SendEvent(context.Background(), "friendship-events",
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

func (f *FriendshipService) BlockUser(ctx context.Context, blockerId, blockedId int64, reason string) error {
	f.log.WithFields(logrus.Fields{
		"blocker_id": blockerId,
		"blocked_id": blockedId,
		"reason":     reason,
	}).Info("Blocking user")

	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		existingBlock, err := tx.GetBlock(ctx, blockerId, blockedId)
		if err != nil && !errors.Is(err, repository.ErrBlockNotFound) {
			return fmt.Errorf("failed to check existing block: %w", err)
		}
		if existingBlock != nil {
			return errors.New("user already blocked")
		}

		if err := tx.DeleteFriend(ctx, blockerId, blockedId); err != nil &&
			!errors.Is(err, errors.New("friendship not found")) {
			f.log.WithError(err).Warn("Failed to delete friend relationship")
		}

		if err := tx.CancelPendingRequestBetweenUsers(ctx, blockerId, blockedId); err != nil {
			f.log.WithError(err).Warn("Failed to cancel pending requests")
		}

		block := &models.BlockedUser{
			BlockerId: blockerId,
			BlockedId: blockedId,
			Reason:    reason,
		}
		if err := tx.CreateBlock(ctx, block); err != nil {
			f.log.WithError(err).Error("Failed to create block")
			return fmt.Errorf("failed to block user: %w", err)
		}

		history := &models.FriendshipHistory{
			EventType: "blocked",
			UserId:    blockerId,
			TargetId:  blockedId,
			NewStatus: stringPtr("blocked"),
			Metadata:  createMetadata("reason", reason),
		}
		if err := tx.CreateHistory(ctx, history); err != nil {
			f.log.WithError(err).Warn("Failed to create history entry (non-critical)")
		}

		event := kafka.NewBlockEvent(blockerId, blockedId, "block", reason)
		go func() {
			if err := f.outboxKafkaProducer.SendEvent(context.Background(), "friendship-events",
				fmt.Sprintf("%d", blockerId), event); err != nil {
				f.log.WithError(err).Error("Failed to send Kafka event")
			}
		}()

		f.log.WithFields(logrus.Fields{
			"blocker_id": blockerId,
			"blocked_id": blockedId,
		}).Info("User blocked successfully")

		return nil
	})
}

func (f *FriendshipService) UnblockUser(ctx context.Context, blockerId, blockedId int64) error {
	f.log.WithFields(logrus.Fields{
		"blocker_id": blockerId,
		"blocked_id": blockedId,
	}).Info("Unblocking user")

	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		if err := tx.DeleteBlock(ctx, blockerId, blockedId); err != nil {
			if errors.Is(err, errors.New("block not found")) {
				return errors.New("user not blocked")
			}
			return fmt.Errorf("failed to unblock user: %w", err)
		}

		history := &models.FriendshipHistory{
			EventType: "unblocked",
			UserId:    blockerId,
			TargetId:  blockedId,
			OldStatus: stringPtr("blocked"),
			NewStatus: stringPtr("unblocked"),
		}
		if err := tx.CreateHistory(ctx, history); err != nil {
			f.log.WithError(err).Warn("Failed to create history entry (non-critical)")
		}

		event := kafka.NewBlockEvent(blockerId, blockedId, "unblock", "")
		go func() {
			if err := f.outboxKafkaProducer.SendEvent(context.Background(), "friendship-events",
				fmt.Sprintf("%d", blockerId), event); err != nil {
				f.log.WithError(err).Error("Failed to send Kafka event")
			}
		}()

		f.log.WithFields(logrus.Fields{
			"blocker_id": blockerId,
			"blocked_id": blockedId,
		}).Info("User unblocked successfully")

		return nil
	})
}

func (f *FriendshipService) DeleteFromFriendList(ctx context.Context, userId, friendId int64) error {
	f.log.WithFields(logrus.Fields{
		"user_id":   userId,
		"friend_id": friendId,
	}).Info("Removing from friend list")

	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		if err := tx.DeleteFriend(ctx, userId, friendId); err != nil {
			if errors.Is(err, errors.New("friendship not found")) {
				return errors.New("users are not friends")
			}
			return fmt.Errorf("failed to delete friend: %w", err)
		}

		history := &models.FriendshipHistory{
			EventType: "unfriended",
			UserId:    userId,
			TargetId:  friendId,
			OldStatus: stringPtr("accepted"),
			NewStatus: stringPtr("unfriended"),
		}
		if err := tx.CreateHistory(ctx, history); err != nil {
			f.log.WithError(err).Warn("Failed to create history entry (non-critical)")
		}

		event := kafka.NewFriendEvent(userId, friendId, "remove")
		go func() {
			if err := f.outboxKafkaProducer.SendEvent(context.Background(), "friendship-events",
				fmt.Sprintf("%d", userId), event); err != nil {
				f.log.WithError(err).Error("Failed to send Kafka event")
			}
		}()

		f.log.WithFields(logrus.Fields{
			"user_id":   userId,
			"friend_id": friendId,
		}).Info("Friend removed successfully")

		return nil
	})
}

func (f *FriendshipService) GetFriendList(ctx context.Context, userId int64) (*service_dto.GetUserViewListResponse, error) {
	f.log.WithField("user_id", userId).Info("Getting friend list")

	friends, err := f.friendshipRepo.GetFriendList(ctx, userId)
	if err != nil {
		f.log.WithError(err).Error("Failed to get friend list")
		return nil, fmt.Errorf("failed to get friend list: %w", err)
	}

	if len(friends) == 0 {
		return &service_dto.GetUserViewListResponse{
			UserList: []*service_dto.GetUserResponse{},
		}, nil
	}

	friendIds := make([]int64, 0, len(friends))
	for _, friend := range friends {
		friendId := friend.FriendId
		if friendId == userId {
			friendId = friend.UserId
		}
		friendIds = append(friendIds, friendId)
	}

	users, err := f.userService.GetUserByIds(ctx, friendIds)
	if err != nil {
		f.log.WithError(err).Error("Failed to get user details")
		return nil, fmt.Errorf("failed to get user details: %w", err)
	}

	userViews := make([]*service_dto.GetUserResponse, 0, len(users))
	for _, user := range users {
		userViews = append(userViews, &service_dto.GetUserResponse{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		})
	}

	f.log.WithFields(logrus.Fields{
		"user_id":       userId,
		"friends_count": len(userViews),
	}).Info("Friend list retrieved successfully")

	return &service_dto.GetUserViewListResponse{
		UserList: userViews,
	}, nil
}

func (f *FriendshipService) CheckRequestState(ctx context.Context, userId, targetId int64) (string, error) {
	f.log.WithFields(logrus.Fields{
		"user_id":   userId,
		"target_id": targetId,
	}).Info("Checking request state")

	request, err := f.friendshipRepo.GetActiveRequestBetweenUsers(ctx, userId, targetId)
	if err != nil {
		if errors.Is(err, errors.New("friendship not found")) {
			return "none", nil
		}
		f.log.WithError(err).Error("Failed to check request state")
		return "", fmt.Errorf("failed to check request state: %w", err)
	}

	status := request.Status
	f.log.WithFields(logrus.Fields{
		"user_id":   userId,
		"target_id": targetId,
		"status":    status,
	}).Info("Request state checked")

	return status, nil
}

func createMetadata(key, value string) json.RawMessage {
	metadata := map[string]string{key: value}
	data, _ := json.Marshal(metadata)
	return data
}

func stringPtr(s string) *string {
	return &s
}
