package repository

import (
	"chat_service/internal/presence/repository/repo_dto"
	"context"
	"time"
)

type PresenceRepo interface {
	AddConnection(ctx context.Context, userId, connId int64, device string) error
	RemoveConnection(ctx context.Context, userId, connId int64) error
	TouchConnection(ctx context.Context, connId int64) error

	GetUserConnections(ctx context.Context, userId int64) ([]repo_dto.Connection, error)
	CleanupDanglingConnections(ctx context.Context, userId int64) error

	SetLastSeen(ctx context.Context, userID int64, t time.Time) error
	GetLastSeen(ctx context.Context, userID int64) (time.Time, error)
}
