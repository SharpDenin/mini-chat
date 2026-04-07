package service

import (
	"context"
	"github.com/sirupsen/logrus"
	"profile_service/internal/kafka"
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
		// ... бизнес-логика ...

		// Отправляем событие
		event := kafka.NewFriendRequestSentEvent(senderId, receiverId, senderId /*request.Id */, message)

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
