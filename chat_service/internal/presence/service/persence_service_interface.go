package service

import (
	sDto "chat_service/internal/presence/service/dto"
	"context"
	"time"
)

type PresenceServiceInterface interface {
	MarkOnline(ctx context.Context, userId int64, opts ...sDto.MarkOptionRequest) error
	MarkOffline(ctx context.Context, userId int64, opts ...sDto.MarkOptionRequest) error
	UpdateLastSeen(ctx context.Context, userId int64) error

	GetPresence(ctx context.Context, userId int64) (*sDto.PresenceResponse, error)
	GetBulkPresence(ctx context.Context, userIds []int64) (*sDto.BulkPresenceResponse, error)
	GetOnlineUsers(ctx context.Context, userIds []int64) ([]int64, error)
	GetRecentlyOnline(ctx context.Context, since time.Time) ([]int64, error)

	AddConnection(ctx context.Context, userId int64, connId int64, deviceType string) error
	RemoveConnection(ctx context.Context, userId int64, connId int64) error
	GetUserConnections(ctx context.Context, userId int64) ([]int64, error)

	CleanupStaleData(ctx context.Context) error
	HealthCheck(ctx context.Context) error

	SubscribeStatusChanges(ctx context.Context) (<-chan sDto.StatusChangeEvent, error)
}
