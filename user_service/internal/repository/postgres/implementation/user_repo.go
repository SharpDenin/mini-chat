package implementation

import (
	"context"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"user_service/internal/app/auth/dto"
	"user_service/internal/app/auth/model"
	"user_service/internal/repository/postgres"
)

type UserRepo struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewUserRepo(db *gorm.DB, log *logrus.Logger) postgres.UserRepoInterface {
	return &UserRepo{
		db:  db,
		log: log,
	}

}

func (u UserRepo) Create(ctx context.Context, person *model.User) (*dto.UserCreateDTO, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserRepo) GetById(ctx context.Context, id int64) (*dto.UserViewDTO, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserRepo) GetAll(ctx context.Context, limit, offset int) ([]*dto.UserViewDTO, int, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserRepo) Update(ctx context.Context, id int64, person *model.User) (*dto.UserViewDTO, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserRepo) Delete(ctx context.Context, id int64) error {
	//TODO implement me
	panic("implement me")
}
