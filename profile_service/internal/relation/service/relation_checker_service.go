package service

import (
	"context"
	"net/http"
	"profile_service/internal/relation/repository"
	"profile_service/internal/relation/service/helpers"
	"profile_service/internal/relation/service/interfaces"
	"profile_service/internal/user/service"
	"profile_service/middleware_profile"
)

type RelationChecker struct {
	userService          service.UserServiceInterface
	friendshipRepository repository.FriendshipRepositoryInterface
}

func NewRelationChecker(userService service.UserServiceInterface,
	friendshipRepository repository.FriendshipRepositoryInterface) interfaces.UserRelationCheckerInterface {
	return &RelationChecker{
		userService:          userService,
		friendshipRepository: friendshipRepository,
	}
}

func (r *RelationChecker) CheckUsersAreFriends(ctx context.Context, a, b int64) (bool, error) {
	return r.friendshipRepository.AreFriends(ctx, a, b)
}

func (r *RelationChecker) CheckUserIsBlocked(ctx context.Context, to int64) (bool, error) {
	from, err := helpers.GetUserIdFromContext(ctx)
	if err != nil {
		return false, middleware_profile.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}
	return r.friendshipRepository.IsBlocked(ctx, from, to)
}
