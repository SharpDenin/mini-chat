package service

import (
	"chat_service/internal/models"
	"chat_service/internal/repository/room_repo"
	"chat_service/pkg/grpc_client"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"proto/generated/profile"
	"proto/middleware"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RoomService struct {
	profileClient *grpc_client.ProfileClient
	rRepo         room_repo.RoomRepoInterface
	log           *logrus.Logger
}

func NewRoomService(profileClient *grpc_client.ProfileClient, rRepo room_repo.RoomRepoInterface,
	log *logrus.Logger) RoomServiceInterface {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &RoomService{
		profileClient: profileClient,
		rRepo:         rRepo,
		log:           log,
	}
}

func (r *RoomService) CreateRoom(ctx context.Context, name string) (int64, error) {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		r.log.WithFields(logrus.Fields{
			"error": "context is not a gin context",
		}).Warn("Failed to get gin context")
		return 0, middleware.NewCustomError(http.StatusInternalServerError, "context is not a gin context", errors.New("context is not a gin context"))
	}
	userId, exists := ginCtx.Get("user_id")
	if !exists {
		r.log.WithFields(logrus.Fields{
			"error": "user_id is not a string",
		}).Warn("Failed to convert user_id to string")
		return 0, middleware.NewCustomError(http.StatusUnauthorized, "user_id not found in context", nil)
	}
	userIdStr, ok := userId.(string)
	if !ok {
		r.log.Warn("Error while converting user id to string")
		return 0, middleware.NewCustomError(http.StatusInternalServerError, "user_id is not a string", nil)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	userReq := &profile.UserExistsRequest{UserId: userIdStr}
	userResp, err := r.profileClient.UserExists(ctx, userReq)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"user_id": userIdStr,
			"error":   err,
		}).Warn("Failed to check user existence")
		return 0, middleware.NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to check user existence: %v", err), err)
	}
	if !userResp.Exists {
		r.log.WithFields(logrus.Fields{
			"user_id": userIdStr,
		}).Warn("User does not exist")
		return 0, middleware.NewCustomError(http.StatusNotFound, "user does not exist", nil)
	}

	if strings.TrimSpace(name) == "" {
		r.log.WithFields(logrus.Fields{
			"name": name,
		}).Warn("Room name is empty")
		return 0, middleware.NewCustomError(http.StatusBadRequest, "room name cannot be empty", nil)
	}

	room := &models.Room{
		Name: name,
	}

	if err := r.rRepo.Create(ctx, room); err != nil {
		r.log.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Warn("Failed to create room")
		return 0, middleware.NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create room: %v", err), err)
	}

	r.log.WithFields(logrus.Fields{
		"room_id": room.Id,
		"name":    name,
	}).Info("Room created successfully")

	return room.Id, nil
}

func (r *RoomService) RenameRoom(ctx context.Context, roomID int64, newName string) error {
	//TODO implement me
	panic("implement me")
}

func (r *RoomService) DeleteRoom(ctx context.Context, roomID int64) error {
	//TODO implement me
	panic("implement me")
}

func (r *RoomService) GetRoom(ctx context.Context, roomID int64) (*models.Room, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RoomService) ListRooms(ctx context.Context, search string, limit, offset int) ([]*models.Room, error) {
	//TODO implement me
	panic("implement me")
}
