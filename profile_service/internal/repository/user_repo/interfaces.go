package user_repo

import (
	"context"
	uModel "profile_service/internal/app/user/models"
	"profile_service/internal/app/user/service/dto"
)

type UserRepoInterface interface {
	Create(ctx context.Context, person *uModel.User) (*uModel.User, error)
	GetById(ctx context.Context, id int64) (*uModel.User, error)
	GetAll(ctx context.Context, filter dto.SearchUserFilter) (int, []*uModel.User, error)
	Update(ctx context.Context, id int64, person *uModel.User) (*uModel.User, error)
	Delete(ctx context.Context, id int64) error
}
