package service

import (
	"context"
	"encoding/json"
	"errors"
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
	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		// Проверяем блокировку
		blocked, err := tx.IsBlocked(ctx, senderId, receiverId)
		if err != nil {
			return err
		}
		if blocked {
			return errors.New("cannot send friend request: user is blocked")
		}

		// Проверяем дружбу
		areFriends, err := tx.AreFriends(ctx, senderId, receiverId)
		if err != nil {
			return err
		}
		if areFriends {
			return errors.New("users are already friends")
		}

		// Проверяем активный запрос
		existingRequest, err := tx.GetActiveRequestBetweenUsers(ctx, senderId, receiverId)
		if err == nil && existingRequest != nil {
			return errors.New("friend request already exists")
		}

		// Создаем запрос
		request := &models.FriendRequest{
			SenderId:   senderId,
			ReceiverId: receiverId,
			Status:     "pending",
			Message:    message,
		}

		if err := tx.CreateFriendRequest(ctx, request); err != nil {
			return err
		}

		// Записываем в историю
		history := &models.FriendshipHistory{
			EventType: "request_sent",
			UserId:    senderId,
			TargetId:  receiverId,
			NewStatus: stringPtr("pending"),
			RequestId: &request.Id,
			Metadata:  createMetadata("message", message),
		}

		if err := tx.CreateHistory(ctx, history); err != nil {
			return err
		}

		// Отправляем событие
		event := kafka.NewFriendRequestSentEvent(senderId, receiverId, request.Id, message)

		// Отправляем через Kafka producer
		go f.kafkaProducer.SendEvent(context.Background(), "friendship-events",
			string(senderId), event)

		return nil
	})
}

func (f FriendshipService) AnswerFriendRequest(ctx context.Context, requestId, userId int64, accept bool) error {
	//TODO implement me
	panic("implement me")
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
