package service

import "context"

type UserRelationChecker interface {
	AreFriends(ctx context.Context, a, b int64) (bool, error)
	IsBlocked(ctx context.Context, from, to int64) (bool, error)
}
