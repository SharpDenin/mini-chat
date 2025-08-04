package service

import (
	"context"
	"user_service/internal/app/user/service/dto"
)

type UserServiceInterface interface {
	GetUserById(ctx context.Context, userId int64) (*dto.GetUserResponse, error)
	GetAllUsers(ctx context.Context, filter dto.SearchUserFilter) (*dto.GetUserViewListResponse, error)
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (int64, error)
	UpdateUser(ctx context.Context, id int64, req *dto.UpdateUserRequest) error
	DeleteUser(ctx context.Context, userId int64) error

	//GetAuthorizedUsers(ctx context.Context, limit, offset int) (int, []*model.User, error)
	//GetUserAuthTokens(ctx context.Context, userId int64) ([]string, error)
}
