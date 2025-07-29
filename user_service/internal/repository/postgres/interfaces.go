package postgres

import (
	"context"
	"user_service/internal/app/auth/dto"
	"user_service/internal/app/auth/model"
)

type UserRepoInterface interface {
	Create(ctx context.Context, person *model.User) (*dto.UserCreateDTO, error)
	GetById(ctx context.Context, id int64) (*dto.UserViewDTO, error)
	GetAll(ctx context.Context, limit, offset int) ([]*dto.UserViewDTO, int, error)
	Update(ctx context.Context, id int64, person *model.User) (*dto.UserViewDTO, error)
	Delete(ctx context.Context, id int64) error
}

//type Tx interface {
//	User() UserRepoInterface
//	DB() *gorm.DB
//}
//
//type TxOptions struct {
//	IsolationLevel sql.IsolationLevel
//	ReadOnly       bool
//}
//
//type TxManager interface {
//	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx Tx) error, opts ...TxOptions) error
//}
