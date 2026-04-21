package interfaces

import (
	"context"
	"profile_service/internal/user/service/service_dto"
)

type FriendshipServiceInterface interface {
	SendFriendRequest(ctx context.Context, receiverId int64, message string) error
	AnswerFriendRequest(ctx context.Context, requestId int64, accept bool) error
	CancelFriendRequest(ctx context.Context, requestId int64) error
	BlockUser(ctx context.Context, blockedId int64, reason string) error
	UnblockUser(ctx context.Context, blockedId int64) error
	DeleteFromFriendList(ctx context.Context, friendId int64) error
	GetFriendList(ctx context.Context) (*service_dto.GetUserViewListResponse, error)
	CheckRequestState(ctx context.Context, targetId int64) (string, error)
}
