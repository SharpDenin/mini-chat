package service

import (
	"context"
	"profile_service/internal/relation/service/interfaces"
	"profile_service/internal/user/service"
)

type RelationChecker struct {
	userService service.UserServiceInterface
}

func NewRelationChecker(userService service.UserServiceInterface) interfaces.UserRelationCheckerInterface {
	return &RelationChecker{
		userService: userService,
	}
}

func (r *RelationChecker) CheckUsersAreFriends(ctx context.Context, a, b int64) (bool, error) {
	// TODO: заменить на реальную проверку дружбы
	return true, nil
}

func (r *RelationChecker) CheckUserIsBlocked(ctx context.Context, from, to int64) (bool, error) {
	// TODO: заменить на реальную проверку блокировки
	return false, nil
}
