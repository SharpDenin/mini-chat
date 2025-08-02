package service

import (
	"context"
	"user_service/internal/app/user/model"
)

type UserServiceInterface interface {
	GetUserById(ctx context.Context, userId int64) (*model.User, error)
	GetAllUsers(ctx context.Context, limit, offset int) (int, []*model.User, error)
	SearchUser(ctx context.Context, username string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) (int64, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, userId int64) error

	GetAuthorizedUsers(ctx context.Context, limit, offset int) (int, []*model.User, error)
	GetUserAuthTokens(ctx context.Context, userId int64) ([]string, error)
}
