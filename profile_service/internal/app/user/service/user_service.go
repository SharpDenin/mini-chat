package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"profile_service/internal/app/user/models"
	"profile_service/internal/app/user/service/dto"
	"profile_service/internal/app/user/service/helpers"
	"profile_service/internal/repository/user_repo"
	"profile_service/internal/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserService struct {
	uRepo user_repo.UserRepoInterface
	log   *logrus.Logger
}

func NewUserService(uRepo user_repo.UserRepoInterface, log *logrus.Logger) UserServiceInterface {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &UserService{
		uRepo: uRepo,
		log:   log,
	}
}

func (u *UserService) GetUserById(ctx context.Context, userId int64) (*dto.GetUserResponse, error) {
	u.log.Debugf("GetUserById %v", userId)
	if err := helpers.ValidateUserId(userId); err != nil {
		return nil, utils.NewCustomError(http.StatusBadRequest, fmt.Sprintf("Validation error %v", userId), err)
	}
	user, err := u.uRepo.GetById(ctx, userId)
	if err != nil {
		return nil, u.handleError(err, userId, "GetUserById")
	}
	response := &dto.GetUserResponse{
		Id:        user.Id,
		Name:      user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
	return response, nil
}

func (u *UserService) GetAllUsers(ctx context.Context, filter dto.SearchUserFilter) (*dto.GetUserViewListResponse, error) {
	u.log.Debugf("GetAllUsers")
	if filter.Limit > 50 {
		return nil, utils.NewCustomError(http.StatusBadRequest, "Limit should be between 0 and 50", nil)
	}
	if filter.Offset < 0 {
		return nil, utils.NewCustomError(http.StatusBadRequest, "Offset cannot be negative", nil)
	}
	total, users, err := u.uRepo.GetAll(ctx, filter)
	if err != nil {
		return nil, u.handleError(err, 0, "GetAllUsers")
	}
	response := &dto.GetUserViewListResponse{
		UserList: make([]*dto.GetUserResponse, 0, len(users)),
		Limit:    filter.Limit,
		Offset:   filter.Offset,
		Total:    total,
	}
	for _, user := range users {
		response.UserList = append(response.UserList, &dto.GetUserResponse{
			Id:        user.Id,
			Name:      user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		})
	}
	return response, nil
}

func (u *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (int64, error) {
	u.log.Debugf("CreateUser")
	if err := helpers.ValidateUserForCreate(req); err != nil {
		return 0, utils.NewCustomError(http.StatusBadRequest, "Validation error", err)
	}
	userModel := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}
	uModel, err := u.uRepo.Create(ctx, userModel)
	if err != nil {
		return 0, u.handleError(err, 0, "CreateUser")
	}
	return uModel.Id, nil
}

func (u *UserService) UpdateUser(ctx context.Context, userId int64, req *dto.UpdateUserRequest) error {
	u.log.Debugf("UpdateUser")
	currentUser, err := u.uRepo.GetById(ctx, userId)
	if err != nil {
		return u.handleError(err, userId, "UpdateUser")
	}
	if req.Username != nil {
		currentUser.Username = *req.Username
	}
	if req.Email != nil {
		currentUser.Email = *req.Email
	}
	if req.Password != nil {
		currentUser.Password = *req.Password
	}
	if err := helpers.ValidateUserForUpdate(currentUser); err != nil {
		return utils.NewCustomError(http.StatusBadRequest, "Validation error", err)
	}
	updatedUser, err := u.uRepo.Update(ctx, userId, currentUser)
	if err != nil {
		return u.handleError(err, userId, "UpdateUser")
	}
	if updatedUser == nil {
		return utils.NewCustomError(http.StatusNotFound, "User not found", nil)
	}
	return nil
}

func (u *UserService) DeleteUser(ctx context.Context, userId int64) error {
	u.log.Debugf("DeleteUser")
	if err := helpers.ValidateUserId(userId); err != nil {
		return utils.NewCustomError(http.StatusBadRequest, fmt.Sprintf("Validation error %v", userId), err)
	}
	err := u.uRepo.Delete(ctx, userId)
	if err != nil {
		return u.handleError(err, userId, "DeleteUser")
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

func (u *UserService) handleError(err error, id int64, operation string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		u.log.Infof("User Not Found, id: %d", id)
		return utils.NewCustomError(http.StatusNotFound, fmt.Sprintf("User not found, id: %d", id), err)
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return utils.NewCustomError(http.StatusConflict, fmt.Sprintf("User already exists in %s", operation), err)
	}
	return utils.NewCustomError(http.StatusInternalServerError, fmt.Sprintf("%s error", operation), err)
}
