package service

import "context"

type RelationChecker struct {
	userService UserServiceInterface
}

func NewRelationChecker(userService UserServiceInterface) *RelationChecker {
	return &RelationChecker{
		userService: userService,
	}
}

func (r *RelationChecker) AreFriends(ctx context.Context, a, b int64) (bool, error) {
	// TODO: заменить на реальную проверку дружбы
	return true, nil
}

func (r *RelationChecker) IsBlocked(ctx context.Context, from, to int64) (bool, error) {
	// TODO: заменить на реальную проверку блокировки
	return false, nil
}
