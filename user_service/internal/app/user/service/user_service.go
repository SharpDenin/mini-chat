package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"os"
	"user_service/internal/app/user/entities/dto"
	"user_service/internal/app/user/entities/model"
	"user_service/internal/app/user/validation"
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
	if err := validation.ValidateUserId(userId); err != nil {
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

func (u *UserService) GetAllUsers(ctx context.Context, filter dto.SearchUserFilterDTO) (int, []*model.User, error) {
	u.log.Debugf("GetAllUsers")
	if filter.Limit > 50 || filter.Limit <= 0 {
		u.log.WithError(fmt.Errorf("invalid limit: %v", filter.Limit))
		return 0, nil, fmt.Errorf("limit should be between 0 and 50")
	}
	if filter.Offset < 0 {
		u.log.WithError(fmt.Errorf("invalid offset: %v", filter.Offset))
		return 0, nil, fmt.Errorf("offset cannot be negative")
	}

	total, users, err := u.uRepo.GetAll(ctx, filter)
	if err != nil {
		u.log.Errorf("GetAllUsers error: %v", err)
		return 0, nil, fmt.Errorf("GetAll error: %v", err)
	}
	return total, users, nil
}

func (u *UserService) CreateUser(ctx context.Context, user *model.User) (int64, error) {
	u.log.Debugf("CreateUser")
	if err := validation.ValidateUserForCreate(user); err != nil {
		u.log.Errorf("userModel validation error %v: %v", user, err)
		return 0, fmt.Errorf("validation error %v: %w", user, err)
	}
	uModel, err := u.uRepo.Create(ctx, user)
	if err != nil {
		u.log.Errorf("CreateUser error: %v", err)
		return 0, fmt.Errorf("CreateUser error: %w", err)
	}
	err = validation.ValidateUserForCreate(uModel)
	if err != nil {
		u.log.Warnf("entity validation error %v: %v", uModel, err)
		return uModel.Id, nil
	}
	return uModel.Id, nil
}

func (u *UserService) UpdateUser(ctx context.Context, userId int64, user *model.User) error {
	u.log.Debugf("UpdateUser")
	if err := validation.ValidateUserForUpdate(user); err != nil {
		u.log.Errorf("userModel validation error %v: %v", user, err)
		return fmt.Errorf("validation error %v: %w", user, err)
	}
	uModel, err := u.uRepo.Update(ctx, userId, user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.log.Infof("User Not Found, id: %d", userId)
			return fmt.Errorf("user not found: %w", err)
		}
		u.log.Errorf("UpdateUser error: %v", err)
		return fmt.Errorf("UpdateUser error: %w", err)
	}
	err = validation.ValidateUserForUpdate(uModel)
	if err != nil {
		u.log.Warnf("entity validation error %v: %v", uModel, err)
		return nil
	}
	return nil
}

func (u *UserService) DeleteUser(ctx context.Context, userId int64) error {
	u.log.Debugf("DeleteUser")
	if err := validation.ValidateUserId(userId); err != nil {
		u.log.Errorf("userId validation error %v: %v", userId, err)
		return fmt.Errorf("validation error %v: %w", userId, err)
	}
	err := u.uRepo.Delete(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.log.Infof("User Not Found, id: %d", userId)
			return fmt.Errorf("user not found: %w", err)
		}
		u.log.Errorf("DeleteUser error: %v", err)
		return fmt.Errorf("DeleteUser error: %w", err)
	}
	return nil
}

//func (u *UserService) GetAuthorizedUsers(ctx context.Context, limit, offset int) (int, []*model.User, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (u *UserService) GetUserAuthTokens(ctx context.Context, userId int64) ([]string, error) {
//	//TODO implement me
//	panic("implement me")
//}
