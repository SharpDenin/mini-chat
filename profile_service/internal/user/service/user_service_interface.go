package service

import (
	"context"
	"profile_service/internal/user/service/service_dto"
)

type UserServiceInterface interface {
	GetUserById(ctx context.Context, userId int64) (*service_dto.GetUserResponse, error)
	GetAllUsers(ctx context.Context, filter service_dto.SearchUserFilter) (*service_dto.GetUserViewListResponse, error)
	CreateUser(ctx context.Context, req *service_dto.CreateUserRequest) (int64, error)
	UpdateUser(ctx context.Context, id int64, req *service_dto.UpdateUserRequest) error
	DeleteUser(ctx context.Context, userId int64) error

	GetAllUsersToCheckAuth(ctx context.Context, filter service_dto.SearchUserFilter) (*service_dto.GetUserViewListResponse, error)
	//GetAuthorizedUsers(ctx context.Context, limit, offset int) (int, []*model.User, error)
	//GetUserAuthTokens(ctx context.Context, userId int64) ([]string, error)
}
