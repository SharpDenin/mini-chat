package service

import (
	"context"
	"profile_service/internal/relation/repository"
	"profile_service/internal/relation/service/interfaces"
	"profile_service/internal/user/service"
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

func (r *RelationChecker) CheckUserIsBlocked(ctx context.Context, from, to int64) (bool, error) {
	return r.friendshipRepository.IsBlocked(ctx, from, to)
}
