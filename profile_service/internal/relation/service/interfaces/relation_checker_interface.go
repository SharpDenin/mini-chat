package interfaces

import "context"

type UserRelationCheckerInterface interface {
	CheckUsersAreFriends(ctx context.Context, a, b int64) (bool, error)
	CheckUserIsBlocked(ctx context.Context, from, to int64) (bool, error)
}
