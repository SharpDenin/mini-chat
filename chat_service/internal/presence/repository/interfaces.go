package repository

import (
	rModels "chat_service/internal/presence/repository/repo_dto"
	"context"
	"time"
)

type PresenceRepoInterface interface {
	SetOnline(ctx context.Context, userId string) error
	SetOffline(ctx context.Context, userId string) error
	SetLastSeen(ctx context.Context, userId string) error

	IsOnline(ctx context.Context, userId string) (bool, error)
	GetLastSeen(ctx context.Context, userId string) (time.Time, error)
	GetOnlineFriends(ctx context.Context, userId string, friendsIds []string) ([]string, error)
	GetUserPresence(ctx context.Context, userId string) (*rModels.UserPresence, error)

	CleanupStaleOnline(ctx context.Context) error
}
