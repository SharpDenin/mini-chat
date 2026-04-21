package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"profile_service/internal/relation/models"
)

const (
	RequestStatusPending   = "pending"
	RequestStatusAccepted  = "accepted"
	RequestStatusRejected  = "rejected"
	RequestStatusCancelled = "cancelled"
)

var (
	ErrFriendshipNotFound = errors.New("friendship not found")
	ErrBlockNotFound      = errors.New("block not found")
)

type FriendshipRepo struct {
	db  *gorm.DB
	log *logrus.Entry
}

func NewFriendshipRepository(db *gorm.DB, log *logrus.Entry) FriendshipRepositoryInterface {
	return &FriendshipRepo{
		db:  db,
		log: log,
	}
}

func (f *FriendshipRepo) CreateFriendRequest(ctx context.Context, request *models.FriendRequest) error {
	if err := f.db.WithContext(ctx).
		Create(request).Error; err != nil {
		f.log.WithFields(logrus.Fields{"error": err}).Error("Failed to create friend request")

		return fmt.Errorf("create friend request error: %w", err)
	}

	return nil
}

func (f *FriendshipRepo) GetPendingRequest(ctx context.Context, requestId, receiverId int64) (*models.FriendRequest, error) {
	var request models.FriendRequest
	err := f.db.WithContext(ctx).
		Where("id = ? AND receiver_id = ? AND status = ?", requestId, receiverId, RequestStatusPending).
		First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return nil, ErrFriendshipNotFound
		}
		f.log.WithFields(logrus.Fields{"error": err, "request_id": requestId, "receiver_id": receiverId}).
			Error("Failed to get pending request")

		return nil, fmt.Errorf("get pending request error: %w", err)
	}

	return &request, nil
}

func (f *FriendshipRepo) GetActiveRequestBetweenUsers(ctx context.Context, userId1, userId2 int64) (*models.FriendRequest, error) {
	var request models.FriendRequest
	err := f.db.WithContext(ctx).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userId1, userId2, userId2, userId1).
		Where("status =?", RequestStatusPending).
		First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return nil, ErrFriendshipNotFound
		}
		f.log.WithFields(logrus.Fields{"error": err, "user_id1": userId1, "user_id2": userId2}).
			Error("Failed to get active request")

		return nil, fmt.Errorf("get active request error: %w", err)
	}

	return &request, nil
}

func (f *FriendshipRepo) UpdateFriendRequestStatus(ctx context.Context, requestId int64, status string) error {
	result := f.db.WithContext(ctx).
		Model(&models.FriendRequest{}).
		Where("id = ?", requestId).
		Update("status", status)

	if result.Error != nil {
		f.log.WithFields(logrus.Fields{"error": result.Error}).Error("Failed to update friend request status")
		return fmt.Errorf("update friend request status error: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrFriendshipNotFound
	}

	return nil
}

func (f *FriendshipRepo) CancelPendingRequestBetweenUsers(ctx context.Context, userId1, userId2 int64) error {
	result := f.db.WithContext(ctx).
		Model(&models.FriendRequest{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userId1, userId2, userId2, userId1).
		Where("status = ?", RequestStatusPending).
		Update("status", RequestStatusCancelled)

	if result.Error != nil {
		f.log.WithFields(logrus.Fields{"error": result.Error}).Error("Failed to cancel pending friend request")
		return fmt.Errorf("cancel pending friend request error: %w", result.Error)
	}

	return nil
}

func (f *FriendshipRepo) CreateFriend(ctx context.Context, friend *models.Friend) error {
	if err := f.db.WithContext(ctx).
		Create(friend).Error; err != nil {
		f.log.WithFields(logrus.Fields{"error": err}).Error("Failed to create friend")

		return fmt.Errorf("create friend error: %w", err)
	}

	return nil
}

func (f *FriendshipRepo) DeleteFriend(ctx context.Context, userId, friendId int64) error {
	result := f.db.WithContext(ctx).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			userId, friendId, friendId, userId).
		Delete(&models.Friend{})

	if result.Error != nil {
		f.log.WithFields(logrus.Fields{"error": result.Error, "user_id": userId, "friend_id": friendId}).
			Error("Failed to delete friend")

		return fmt.Errorf("delete friend error: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrFriendshipNotFound
	}

	return nil
}

func (f *FriendshipRepo) AreFriends(ctx context.Context, userId1, userId2 int64) (bool, error) {
	var count int64
	err := f.db.WithContext(ctx).
		Model(&models.Friend{}).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			userId1, userId2, userId2, userId1).
		Count(&count).Error

	if err != nil {
		f.log.WithFields(logrus.Fields{"error": err, "user_id1": userId1, "user_id2": userId2}).
			Error("Failed to check users are friends")

		return false, fmt.Errorf("check users are friends error: %w", err)
	}

	return count > 0, nil
}

func (f *FriendshipRepo) GetFriendListWithPagination(ctx context.Context, userId int64, limit, offset int) ([]models.Friend, int64, error) {
	var friends []models.Friend
	var total int64

	query := f.db.WithContext(ctx).Model(&models.Friend{}).
		Where("user_id = ? OR friend_id = ?", userId, userId)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count friends error: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&friends).Error; err != nil {
		return nil, 0, fmt.Errorf("get friend list error: %w", err)
	}

	return friends, total, nil
}

func (f *FriendshipRepo) CreateBlock(ctx context.Context, block *models.BlockedUser) error {
	if err := f.db.WithContext(ctx).
		Create(block).Error; err != nil {
		f.log.WithFields(logrus.Fields{"error": err}).Error("Failed to create block")

		return fmt.Errorf("create block error: %w", err)
	}

	return nil
}

func (f *FriendshipRepo) DeleteBlock(ctx context.Context, blockerId, blockedId int64) error {
	result := f.db.WithContext(ctx).
		Where("blocker_id = ? AND blocked_id = ?", blockerId, blockedId).
		Delete(&models.BlockedUser{})

	if result.Error != nil {
		f.log.WithFields(logrus.Fields{"error": result.Error, "blocker_id": blockerId, "blocked_id": blockedId}).
			Error("Failed to delete block")

		return fmt.Errorf("delete block error: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrBlockNotFound
	}

	return nil
}

func (f *FriendshipRepo) IsBlocked(ctx context.Context, blockerId, blockedId int64) (bool, error) {
	var count int64
	err := f.db.WithContext(ctx).
		Model(&models.BlockedUser{}).
		Where("blocker_id = ? AND blocked_id = ?", blockerId, blockedId).
		Count(&count).Error

	if err != nil {
		f.log.WithFields(logrus.Fields{"error": err, "blocker_id": blockerId, "blocked_id": blockedId}).
			Error("Failed to check user is blocked")

		return false, fmt.Errorf("check user is blocked: %w", err)
	}

	return count > 0, nil
}

func (f *FriendshipRepo) GetBlock(ctx context.Context, blockerId, blockedId int64) (*models.BlockedUser, error) {
	var block models.BlockedUser
	err := f.db.WithContext(ctx).
		Where("blocker_id = ? AND blocked_id = ?", blockerId, blockedId).
		First(&block).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBlockNotFound
		}
		f.log.WithFields(logrus.Fields{"error": err, "blocker_id": blockerId, "blocked_id": blockedId}).
			Error("Failed to get block")

		return nil, fmt.Errorf("get block error: %w", err)
	}

	return &block, nil
}

func (f *FriendshipRepo) GetPendingRequestBySender(ctx context.Context, requestId, senderId int64) (*models.FriendRequest, error) {
	var request models.FriendRequest
	err := f.db.WithContext(ctx).
		Where("id = ? AND sender_id = ? AND status = ?", requestId, senderId, RequestStatusPending).
		First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFriendshipNotFound
		}
		return nil, fmt.Errorf("get pending request by sender error: %w", err)
	}
	return &request, nil
}

func (f *FriendshipRepo) CreateHistory(ctx context.Context, history *models.FriendshipHistory) error {
	if err := f.db.WithContext(ctx).
		Create(history).Error; err != nil {
		f.log.WithFields(logrus.Fields{"error": err}).Error("Failed to create history")

		return fmt.Errorf("create history error: %w", err)
	}

	return nil
}
