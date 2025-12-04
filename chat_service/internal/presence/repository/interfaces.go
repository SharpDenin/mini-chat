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
	GetUserPresence(ctx context.Context, userId int64) (*rModels.UserPresence, error)

	CleanupStaleOnline(ctx context.Context) error
}
