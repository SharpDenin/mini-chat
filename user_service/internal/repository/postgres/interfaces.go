package postgres

import (
	"context"
	"user_service/internal/app/auth/model"
)

type UserRepoInterface interface {
	Create(ctx context.Context, person *model.User) (*model.User, error)
	GetById(ctx context.Context, id int64) (*model.User, error)
	GetAll(ctx context.Context, limit, offset int) (int, []*model.User, error)
	Update(ctx context.Context, id int64, person *model.User) (*model.User, error)
	Delete(ctx context.Context, id int64) error
}
