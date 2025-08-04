package service

import (
	"context"
	"user_service/internal/app/user/entities/dto"
	"user_service/internal/app/user/entities/model"
)

type UserServiceInterface interface {
	GetUserById(ctx context.Context, userId int64) (*model.User, error)
	GetAllUsers(ctx context.Context, filter dto.SearchUserFilterDTO) (int, []*model.User, error)
	CreateUser(ctx context.Context, user *model.User) (int64, error)
	UpdateUser(ctx context.Context, id int64, user *model.User) error
	DeleteUser(ctx context.Context, userId int64) error

	//GetAuthorizedUsers(ctx context.Context, limit, offset int) (int, []*model.User, error)
	//GetUserAuthTokens(ctx context.Context, userId int64) ([]string, error)
}
