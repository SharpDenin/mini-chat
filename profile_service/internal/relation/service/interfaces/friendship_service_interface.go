package interfaces

import (
	"context"
	"profile_service/internal/user/service/service_dto"
)

type FriendshipServiceInterface interface {
	SendFriendRequest(ctx context.Context, senderId, receiverId int64, message string) error
	AnswerFriendRequest(ctx context.Context, requestId, userId int64, accept bool) error
	BlockUser(ctx context.Context, blockerId, blockedId int64, reason string) error
	UnblockUser(ctx context.Context, blockerId, blockedId int64) error
	DeleteFromFriendList(ctx context.Context, userId, friendId int64) error
	GetFriendList(ctx context.Context, userId int64) (*service_dto.GetUserViewListResponse, error)
	CheckRequestState(ctx context.Context, userId, targetId int64) (string, error)
}
