package presence_repository

import (
	"context"
	"time"
)

type PresenceRepo interface {
	SetOnline(ctx context.Context, userId string) error
	SetOffline(ctx context.Context, userId string) error
	SetLastSeen(ctx context.Context, userId string) error

	IsOnline(ctx context.Context, userId string) (bool, error)
	GetLastSeen(ctx context.Context, userId string) (time.Time, error)
}
