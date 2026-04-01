package repository

import (
	"context"
	"profile_service/internal/relation/models"
)

type FriendshipRepositoryInterface interface {
	// FriendRequest
	CreateFriendRequest(ctx context.Context, request *models.FriendRequest) error
	GetPendingRequest(ctx context.Context, requestId, receiverId int64) (*models.FriendRequest, error)
	GetActiveRequestBetweenUsers(ctx context.Context, userId1, userId2 int64) (*models.FriendRequest, error)
	UpdateFriendRequestStatus(ctx context.Context, requestId int64, status string) error
	CancelPendingRequestBetweenUsers(ctx context.Context, userId1, userId2 int64) error

	// Friend
	CreateFriend(ctx context.Context, friend *models.Friend) error
	DeleteFriend(ctx context.Context, userId, friendId int64) error
	AreFriends(ctx context.Context, userId1, userId2 int64) (bool, error)
	GetFriendList(ctx context.Context, userId int64) ([]models.Friend, error)

	// Block
	CreateBlock(ctx context.Context, block *models.BlockedUser) error
	DeleteBlock(ctx context.Context, blockerId, blockedId int64) error
	IsBlocked(ctx context.Context, blockerId, blockedId int64) (bool, error)
	GetBlock(ctx context.Context, blockerId, blockedId int64) (*models.BlockedUser, error)

	// History
	CreateHistory(ctx context.Context, history *models.FriendshipHistory) error
}
