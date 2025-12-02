package service

import (
	sDto "chat_service/internal/presence/service/dto"
	"context"
	"time"
)

type PresenceServiceInterface interface {
	MarkOnline(ctx context.Context, userId string, opt ...sDto.MarkOptionRequest) error
	MarkOffline(ctx context.Context, userId string, opt ...sDto.MarkOptionRequest) error
	UpdateLastSeen(ctx context.Context, userId string) error

	GetPresence(ctx context.Context, userId string) (*sDto.PresenceResponse, error)
	GetBulkPresence(ctx context.Context, userIds []string) (*sDto.BulkPresenceResponse, error)
	GetOnlineUsers(ctx context.Context, userIds []string) ([]string, error)
	GetRecentlyOnline(ctx context.Context, since time.Time) ([]string, error)

	AddConnection(ctx context.Context, userId string, connId string, deviceType string) error
	RemoveConnection(ctx context.Context, userId string, connId string) error
	GetUserConnections(ctx context.Context, userId string) ([]string, error)

	CleanupStaleData(ctx context.Context) error
	HealthCheck(ctx context.Context) error

	SubscribeStatusChanges(ctx context.Context) (<-chan sDto.StatusChangeEvent, error)
}
