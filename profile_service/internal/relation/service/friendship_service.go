package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"profile_service/internal/kafka"
	"profile_service/internal/relation/models"
	"profile_service/internal/relation/repository"
	"profile_service/internal/relation/service/helpers"
	"profile_service/internal/relation/service/interfaces"
	"profile_service/internal/user/service"
	"profile_service/internal/user/service/service_dto"
	"profile_service/middleware_profile"

	"github.com/sirupsen/logrus"
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

func (f *FriendshipService) SendFriendRequest(ctx context.Context, receiverId int64, message string) error {
	senderId, err := helpers.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware_profile.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}

	_, err = f.userService.GetUserById(ctx, receiverId)
	if err != nil {
		return helpers.ErrUserNotFound
	}

	f.log.WithFields(logrus.Fields{
		"sender_id":   senderId,
		"receiver_id": receiverId,
	}).Info("Sending friend request")

	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		if receiverId == senderId {
			f.log.WithError(err).Error("Failed to send friend request to yourself")
			return helpers.ErrCannotFriendYourself
		}

		blocked, err := tx.IsBlocked(ctx, senderId, receiverId)
		if err != nil {
			f.log.WithError(err).Error("Failed to check block status")
			return fmt.Errorf("failed to check block status: %w", err)
		}
		if blocked {
			f.log.Warn("User is blocked, cannot send friend request")
			return helpers.ErrBlockedByUser
		}

		areFriends, err := tx.AreFriends(ctx, senderId, receiverId)
		if err != nil {
			f.log.WithError(err).Error("Failed to check friendship status")
			return fmt.Errorf("failed to check friendship status: %w", err)
		}
		if areFriends {
			f.log.Warn("Users are already friends")
			return helpers.ErrAlreadyFriends
		}

		existingRequest, err := tx.GetActiveRequestBetweenUsers(ctx, senderId, receiverId)
		if err != nil && !errors.Is(err, repository.ErrFriendshipNotFound) {
			f.log.WithError(err).Error("Failed to check existing request")
			return fmt.Errorf("failed to check existing request: %w", err)
		}
		if existingRequest != nil {
			f.log.Warn("Friend request already exists")
			return helpers.ErrFriendRequestExists
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
		_ = tx.CreateHistory(ctx, history)

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

func (f *FriendshipService) AnswerFriendRequest(ctx context.Context, requestId int64, accept bool) error {
	userId, err := helpers.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware_profile.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}

	f.log.WithFields(logrus.Fields{
		"request_id": requestId,
		"user_id":    userId,
		"accept":     accept,
	}).Info("Answering friend request")
	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		request, err := tx.GetPendingRequest(ctx, requestId, userId)
		if err != nil {
			if errors.Is(err, repository.ErrFriendshipNotFound) {
				f.log.WithError(err).Warn("Pending request not found")
				return helpers.ErrFriendRequestNotFound
			}
			f.log.WithError(err).Error("Failed to get pending request")
			return fmt.Errorf("failed to get friend request: %w", err)
		}
		if request == nil {
			return helpers.ErrFriendRequestNotFound
		}

		if accept {
			if err := tx.UpdateFriendRequestStatus(ctx, requestId, repository.RequestStatusAccepted); err != nil {
				if errors.Is(err, repository.ErrFriendshipNotFound) {
					return helpers.ErrFriendRequestNotFound
				}

				return fmt.Errorf("failed to update request status: %w", err)
			}

			friend := &models.Friend{
				UserId:   request.SenderId,
				FriendId: request.ReceiverId,
			}
			if err := tx.CreateFriend(ctx, friend); err != nil {
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
			_ = tx.CreateHistory(ctx, history)

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
				if errors.Is(err, repository.ErrFriendshipNotFound) {
					return helpers.ErrFriendRequestNotFound
				}

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
			_ = tx.CreateHistory(ctx, history)

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

func (f *FriendshipService) CancelFriendRequest(ctx context.Context, requestId int64) error {
	userId, err := helpers.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware_profile.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}

	f.log.WithFields(logrus.Fields{
		"request_id": requestId,
		"user_id":    userId,
	}).Info("Cancelling friend request")

	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		request, err := tx.GetPendingRequestBySender(ctx, requestId, userId)
		if err != nil {
			if errors.Is(err, repository.ErrFriendshipNotFound) {
				return helpers.ErrFriendRequestNotFound
			}
			return fmt.Errorf("failed to get pending request: %w", err)
		}

		if err := tx.UpdateFriendRequestStatus(ctx, requestId, repository.RequestStatusCancelled); err != nil {
			return fmt.Errorf("failed to cancel request: %w", err)
		}

		history := &models.FriendshipHistory{
			EventType: "request_cancelled",
			UserId:    userId,
			TargetId:  request.ReceiverId,
			OldStatus: stringPtr(repository.RequestStatusPending),
			NewStatus: stringPtr(repository.RequestStatusCancelled),
			RequestId: &request.Id,
		}
		_ = tx.CreateHistory(ctx, history)

		event := kafka.NewFriendRequestActionEvent(userId, request.ReceiverId, requestId, "cancelled")
		go func() {
			if err := f.outboxKafkaProducer.SendEvent(context.Background(), "friendship-events",
				fmt.Sprintf("%d", userId), event); err != nil {
				f.log.WithError(err).Error("Failed to send Kafka event")
			}
		}()

		f.log.WithFields(logrus.Fields{
			"request_id":  requestId,
			"sender_id":   userId,
			"receiver_id": request.ReceiverId,
		}).Info("Friend request cancelled successfully")

		return nil
	})
}

func (f *FriendshipService) BlockUser(ctx context.Context, blockedId int64, reason string) error {
	blockerId, err := helpers.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware_profile.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}

	_, err = f.userService.GetUserById(ctx, blockedId)
	if err != nil {
		return helpers.ErrUserNotFound
	}

	f.log.WithFields(logrus.Fields{
		"blocker_id": blockerId,
		"blocked_id": blockedId,
		"reason":     reason,
	}).Info("Blocking user")

	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		if blockerId == blockedId {
			f.log.WithError(err).Error("Failed to block yourself")
			return helpers.ErrCannotBlockYourself
		}

		existingBlock, err := tx.GetBlock(ctx, blockerId, blockedId)
		if err != nil && !errors.Is(err, repository.ErrBlockNotFound) {
			return fmt.Errorf("failed to check existing block: %w", err)
		}
		if existingBlock != nil {
			return helpers.ErrUserAlreadyBlocked
		}

		if err := tx.DeleteFriend(ctx, blockerId, blockedId); err != nil && !errors.Is(err, repository.ErrFriendshipNotFound) {
			f.log.WithError(err).Warn("Failed to delete friend relationship (non-critical)")
		}

		if err := tx.CancelPendingRequestBetweenUsers(ctx, blockerId, blockedId); err != nil {
			f.log.WithError(err).Warn("Failed to cancel pending requests (non-critical)")
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
		_ = tx.CreateHistory(ctx, history)

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

func (f *FriendshipService) UnblockUser(ctx context.Context, blockedId int64) error {
	blockerId, err := helpers.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware_profile.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}

	_, err = f.userService.GetUserById(ctx, blockedId)
	if err != nil {
		return helpers.ErrUserNotFound
	}

	f.log.WithFields(logrus.Fields{
		"blocker_id": blockerId,
		"blocked_id": blockedId,
	}).Info("Unblocking user")

	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		if blockerId == blockedId {
			f.log.WithError(err).Error("Failed to unblock yourself")
			return helpers.ErrCannotUnblockYourself
		}

		if err := tx.DeleteBlock(ctx, blockerId, blockedId); err != nil {
			if errors.Is(err, repository.ErrBlockNotFound) {
				return helpers.ErrUserNotBlocked
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
		_ = tx.CreateHistory(ctx, history)

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

func (f *FriendshipService) DeleteFromFriendList(ctx context.Context, friendId int64) error {
	userId, err := helpers.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware_profile.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}

	_, err = f.userService.GetUserById(ctx, friendId)
	if err != nil {
		return helpers.ErrUserNotFound
	}

	f.log.WithFields(logrus.Fields{
		"user_id":   userId,
		"friend_id": friendId,
	}).Info("Removing from friend list")

	return f.txManager.RunInTransaction(ctx, func(tx repository.FriendshipRepositoryInterface) error {
		if userId == friendId {
			f.log.WithError(err).Error("Failed to manage yourself")
			return helpers.ErrCannotDeleteYourself
		}

		if err := tx.DeleteFriend(ctx, userId, friendId); err != nil {
			if errors.Is(err, repository.ErrFriendshipNotFound) {
				return helpers.ErrUsersNotFriends
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
		_ = tx.CreateHistory(ctx, history)

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

func (f *FriendshipService) GetFriendList(ctx context.Context) (*service_dto.GetUserViewListResponse, error) {
	userId, err := helpers.GetUserIdFromContext(ctx)
	if err != nil {
		return nil, middleware_profile.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}

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

func (f *FriendshipService) CheckRequestState(ctx context.Context, targetId int64) (string, error) {
	userId, err := helpers.GetUserIdFromContext(ctx)
	if err != nil {
		return "", middleware_profile.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}

	_, err = f.userService.GetUserById(ctx, targetId)
	if err != nil {
		return "", helpers.ErrUserNotFound
	}

	f.log.WithFields(logrus.Fields{
		"user_id":   userId,
		"target_id": targetId,
	}).Info("Checking request state")

	request, err := f.friendshipRepo.GetActiveRequestBetweenUsers(ctx, userId, targetId)
	if err != nil {
		if errors.Is(err, repository.ErrFriendshipNotFound) {
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
