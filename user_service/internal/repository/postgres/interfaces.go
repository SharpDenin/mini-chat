package postgres

import (
	"context"
	"user_service/internal/app/auth/model"
	"user_service/internal/app/user/dto"
	uModel "user_service/internal/app/user/model"
)

type UserRepoInterface interface {
	Create(ctx context.Context, person *uModel.User) (*uModel.User, error)
	GetById(ctx context.Context, id int64) (*uModel.User, error)
	GetAll(ctx context.Context, filter dto.SearchUserFilterDTO) (int, []*uModel.User, error)
	Update(ctx context.Context, id int64, person *uModel.User) (*uModel.User, error)
	Delete(ctx context.Context, id int64) error
}

type AuthRepoInterface interface {
	SaveToken(ctx context.Context, token *model.AuthToken) error
	GetToken(ctx context.Context, tokenString string) (*model.AuthToken, error)
	RevokeToken(ctx context.Context, tokenString string) error
	RevokeAllUserTokens(ctx context.Context, userId int64) error
	DeleteExpiredTokens(ctx context.Context) error
}
