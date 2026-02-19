package authz

import "context"

type AuthServiceInterface interface {
	CanSendDirect(ctx context.Context, fromUserId, toUserId int64) (bool, error)
}
