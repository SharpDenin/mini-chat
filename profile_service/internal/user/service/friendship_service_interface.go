package service

import "context"

type FriendshipServiceInterface interface {
	SendFriendRequest(ctx context.Context)
	AnswerFriendRequest(ctx context.Context)
	CheckRequestState(ctx context.Context)
	BlockUser(ctx context.Context)
	GetFriendList(ctx context.Context)
	DeleteFromFriendList(ctx context.Context)
}
