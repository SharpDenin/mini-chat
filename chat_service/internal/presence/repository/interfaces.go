package repository

import (
	rModels "chat_service/internal/presence/repository/repo_dto"
	"context"
	"time"
)

type PresenceRepoInterface interface {
	SetOnline(ctx context.Context, userId int64) error
	SetOffline(ctx context.Context, userId int64) error
	SetLastSeen(ctx context.Context, userId int64) error

	IsOnline(ctx context.Context, userId int64) (bool, error)
	GetLastSeen(ctx context.Context, userId int64) (time.Time, error)
	GetOnlineFriends(ctx context.Context, userId int64, friendsIds []int64) ([]int64, error)
	GetUserPresence(ctx context.Context, userId int64) (*rModels.UserPresenceResponse, error)
	GetRecentlyOnline(ctx context.Context, since time.Time) ([]int64, error)

	CleanupStaleOnline(ctx context.Context) error

	AddConnection(ctx context.Context, userId int64, connId int64, deviceType string) error
	RemoveConnection(ctx context.Context, userId int64, connId int64) error
	GetUserConnections(ctx context.Context, userId int64) ([]int64, error)
	GetConnectionInfo(ctx context.Context, userId int64, connId int64) (*rModels.ConnectionInfoResponse, error)
	UpdateConnectionActivity(ctx context.Context, userId int64, connId int64) error
	GetAllUserConnections(ctx context.Context, userId int64) ([]rModels.ConnectionInfoResponse, error)
}
