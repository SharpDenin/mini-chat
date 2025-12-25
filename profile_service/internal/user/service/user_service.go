package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"profile_service/internal/user/models"
	"profile_service/internal/user/repository/profile_repo"
	"profile_service/internal/user/service/helpers"
	"profile_service/internal/user/service/service_dto"
	"profile_service/middleware_profile"
	"profile_service/pkg/grpc_client"
	"profile_service/pkg/grpc_generated/chat"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserService struct {
	presenceClient *grpc_client.PresenceClient
	uRepo          profile_repo.ProfileRepoInterface
	log            *logrus.Logger
}

func NewUserService(presenceClient *grpc_client.PresenceClient, uRepo profile_repo.ProfileRepoInterface, log *logrus.Logger) UserServiceInterface {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &UserService{
		presenceClient: presenceClient,
		uRepo:          uRepo,
		log:            log,
	}
}

func (u *UserService) GetUserById(ctx context.Context, userId int64) (*service_dto.GetUserResponse, error) {
	u.log.Debugf("GetUserById %v", userId)
	if err := helpers.ValidateUserId(userId); err != nil {
		return nil, middleware_profile.NewCustomError(http.StatusBadRequest, fmt.Sprintf("Validation error %v", userId), err)
	}
	user, err := u.uRepo.GetById(ctx, userId)
	if err != nil {
		return nil, u.handleError(err, userId, "GetUserById")
	}

	if user == nil {
		return nil, fmt.Errorf("user with id %d not found", userId)
	}

	serviceStatus, err := u.getUserPresence(ctx, user.Id)
	if err != nil {
		return nil, middleware_profile.NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get user presence: %v", userId), err)
	}

	response := &service_dto.GetUserResponse{
		Id:        user.Id,
		Name:      user.Username,
		Email:     user.Email,
		Status:    serviceStatus,
		CreatedAt: user.CreatedAt,
	}
	return response, nil
}

func (u *UserService) GetAllUsers(ctx context.Context, filter service_dto.SearchUserFilter) (*service_dto.GetUserViewListResponse, error) {
	u.log.Debugf("GetAllUsers")
	if filter.Limit > 50 {
		return nil, middleware_profile.NewCustomError(http.StatusBadRequest, "Limit should be between 0 and 50", nil)
	}
	if filter.Offset < 0 {
		return nil, middleware_profile.NewCustomError(http.StatusBadRequest, "Offset cannot be negative", nil)
	}
	total, users, err := u.uRepo.GetAll(ctx, filter)
	if err != nil {
		return nil, u.handleError(err, 0, "GetAllUsers")
	}
	response := &service_dto.GetUserViewListResponse{
		UserList: make([]*service_dto.GetUserResponse, 0, len(users)),
		Limit:    filter.Limit,
		Offset:   filter.Offset,
		Total:    total,
	}
	for _, user := range users {
		serviceStatus, err := u.getUserPresence(ctx, user.Id)
		if err != nil {
			return nil, middleware_profile.NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get user presence: %v", user.Id), err)
		}

		if filter.Status != "" && filter.Status != serviceStatus {
			continue
		}

		response.UserList = append(response.UserList, &service_dto.GetUserResponse{
			Id:        user.Id,
			Name:      user.Username,
			Email:     user.Email,
			Status:    serviceStatus,
			CreatedAt: user.CreatedAt,
		})
	}

	if filter.Status != "" {
		response.Total = len(response.UserList)
	}

	return response, nil
}

func (u *UserService) GetAllUsersToCheckAuth(ctx context.Context, filter service_dto.SearchUserFilter) (*service_dto.GetUserViewListResponse, error) {
	u.log.Debugf("GetAllUsers")
	if filter.Limit > 50 {
		return nil, middleware_profile.NewCustomError(http.StatusBadRequest, "Limit should be between 0 and 50", nil)
	}
	if filter.Offset < 0 {
		return nil, middleware_profile.NewCustomError(http.StatusBadRequest, "Offset cannot be negative", nil)
	}
	total, users, err := u.uRepo.GetAll(ctx, filter)
	if err != nil {
		return nil, u.handleError(err, 0, "GetAllUsers")
	}
	response := &service_dto.GetUserViewListResponse{
		UserList: make([]*service_dto.GetUserResponse, 0, len(users)),
		Limit:    filter.Limit,
		Offset:   filter.Offset,
		Total:    total,
	}
	for _, user := range users {
		response.UserList = append(response.UserList, &service_dto.GetUserResponse{
			Id:        user.Id,
			Name:      user.Username,
			Email:     user.Email,
			Password:  user.Password,
			CreatedAt: user.CreatedAt,
		})
	}
	return response, nil
}

func (u *UserService) CreateUser(ctx context.Context, req *service_dto.CreateUserRequest) (int64, error) {
	u.log.Debugf("CreateUser")
	if err := helpers.ValidateUserForCreate(req); err != nil {
		return 0, middleware_profile.NewCustomError(http.StatusBadRequest, "Validation error", err)
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

func (u *UserService) UpdateUser(ctx context.Context, userId int64, req *service_dto.UpdateUserRequest) error {
	u.log.Debugf("UpdateUser")
	currentUser, err := u.uRepo.GetById(ctx, userId)
	if err != nil {
		return u.handleError(err, userId, "UpdateUser")
	}

	if currentUser == nil {
		return middleware_profile.NewCustomError(http.StatusNotFound, "User not found", nil)
	}

	if req.Username != nil {
		currentUser.Username = *req.Username
	}
	if req.Email != nil {
		currentUser.Email = *req.Email
	}
	if err := helpers.ValidateUserForUpdate(currentUser); err != nil {
		return middleware_profile.NewCustomError(http.StatusBadRequest, "Validation error", err)
	}
	updatedUser, err := u.uRepo.Update(ctx, userId, currentUser)
	if err != nil {
		return u.handleError(err, userId, "UpdateUser")
	}
	if updatedUser == nil {
		return middleware_profile.NewCustomError(http.StatusNotFound, "User not found", nil)
	}
	return nil
}

func (u *UserService) DeleteUser(ctx context.Context, userId int64) error {
	u.log.Debugf("DeleteUser")
	if err := helpers.ValidateUserId(userId); err != nil {
		return middleware_profile.NewCustomError(http.StatusBadRequest, fmt.Sprintf("Validation error %v", userId), err)
	}
	err := u.uRepo.Delete(ctx, userId)
	if err != nil {
		return u.handleError(err, userId, "DeleteUser")
	}
	return nil
}

func (u *UserService) handleError(err error, id int64, operation string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		u.log.Infof("User Not Found, id: %d", id)
		return middleware_profile.NewCustomError(http.StatusNotFound, fmt.Sprintf("User not found, id: %d", id), err)
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return middleware_profile.NewCustomError(http.StatusConflict, fmt.Sprintf("User already exists in %s", operation), err)
	}
	return middleware_profile.NewCustomError(http.StatusInternalServerError, fmt.Sprintf("%s error", operation), err)
}

func (u *UserService) getUserPresence(ctx context.Context, userId int64) (service_dto.UserStatus, error) {
	if u.presenceClient == nil {
		u.log.WithFields(logrus.Fields{
			"userId": userId,
		}).Debug("Presence client not available, returning unknown status")
		return service_dto.StatusUnknown, nil
	}

	chatReq := &chat.GetPresenceRequest{UserId: userId}

	isOnline, err := helpers.CheckUserPresence(ctx, u.presenceClient, chatReq)
	if err != nil {
		u.log.WithFields(logrus.Fields{
			"userId": userId,
			"error":  err,
		}).Warn("Failed to get userPresence")

		return service_dto.StatusUnknown, nil
	}

	if isOnline {
		return service_dto.StatusOnline, nil
	}

	return service_dto.StatusOffline, nil
}
