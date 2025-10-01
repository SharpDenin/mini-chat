package profile_repo

import (
	"context"
	uModel "profile_service/internal/models"
	"profile_service/internal/service/service_dto"
)

type ProfileRepoInterface interface {
	Create(ctx context.Context, user *uModel.User) (*uModel.User, error)
	GetById(ctx context.Context, id int64) (*uModel.User, error)
	GetAll(ctx context.Context, filter service_dto.SearchUserFilter) (int, []*uModel.User, error)
	Update(ctx context.Context, id int64, user *uModel.User) (*uModel.User, error)
	Delete(ctx context.Context, id int64) error
}
