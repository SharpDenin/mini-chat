package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"math"
	"os"
	"user_service/internal/app/user/model"
	"user_service/internal/repository/postgres/implementation"
)

type UserService struct {
	log   *logrus.Logger
	uRepo implementation.UserRepo
}

func NewUserService(log *logrus.Logger, uRepo implementation.UserRepo) *UserService {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &UserService{
		log:   log,
		uRepo: uRepo,
	}
}

func (u *UserService) GetUserById(ctx context.Context, userId int64) (*model.User, error) {
	u.log.Debugf("GetUserById %v", userId)
	if err := validateUserId(userId); err != nil {
		u.log.Errorf("userId validation error %v: %v", userId, err)
		return nil, fmt.Errorf("validation error %v: %w", userId, err)
	}
	user, err := u.uRepo.GetById(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.log.Infof("User Not Found, id: %d", userId)
			return nil, fmt.Errorf("user not found: %w", err)
		}
		u.log.Errorf("GetUserById error: %v", err)
		return nil, fmt.Errorf("GetUserById error: %w", err)
	}
	return user, nil
}

func (u *UserService) GetAllUsers(ctx context.Context, limit, offset int) (int, []*model.User, error) {
	u.log.Debugf("GetAllUsers")
	if limit > 50 || limit <= 0 {
		u.log.WithError(fmt.Errorf("invalid limit: %v", limit))
		return 0, nil, fmt.Errorf("limit should be between 0 and 50")
	}
	if offset < 0 {
		u.log.WithError(fmt.Errorf("invalid offset: %v", offset))
		return 0, nil, fmt.Errorf("offset cannot be negative")
	}

	total, users, err := u.uRepo.GetAll(ctx, limit, offset)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.log.Infof("No users found")
			return 0, nil, nil
		}
		u.log.Errorf("GetAllUsers error: %v", err)
		return 0, nil, fmt.Errorf("GetAll error: %v", err)
	}
	return total, users, nil
}

// SearchUser TODO  Почитать, как сделать поиск по like
func (u *UserService) SearchUser(ctx context.Context, username string) (*model.User, error) {
	//TODO implement me
	panic("implement me")
}

// CreateUser TODO Валидировать модели приходящие из БД и выводить warning если валидация не прошла
func (u *UserService) CreateUser(ctx context.Context, user *model.User) (int64, error) {
	//TODO А какой ID возвращать?
	panic("implement me")
	//model, err := u.uRepo.Create(ctx, user)
	//return model.Id, nil
}

func (u *UserService) UpdateUser(ctx context.Context, user *model.User) error {
	//TODO implement me
	panic("implement me")
}

func (u *UserService) DeleteUser(ctx context.Context, userId int64) error {
	//TODO implement me
	panic("implement me")
}

func (u *UserService) GetAuthorizedUsers(ctx context.Context, limit, offset int) (int, []*model.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserService) GetUserAuthTokens(ctx context.Context, userId int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func validateUserId(userId int64) error {
	if userId <= 0 {
		return fmt.Errorf("user ID must be positive")
	}
	if userId > math.MaxInt64 {
		return fmt.Errorf("userId is too large")
	}
	return nil
}

func validateUserModel(userModel *model.User) error {
	return nil
}
