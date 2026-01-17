package service

import (
	sDto "chat_service/internal/presence/service/dto"
	"context"
)

type PresenceService interface {
	OnConnect(ctx context.Context, userId, connId int64, device string) error
	OnDisconnect(ctx context.Context, userId, connId int64) error
	OnHeartbeat(ctx context.Context, connId int64) error

	GetPresence(ctx context.Context, userId int64) *sDto.Presence
	GetOnlineFriends(ctx context.Context, userId int64, friends []int64) []int64
}
