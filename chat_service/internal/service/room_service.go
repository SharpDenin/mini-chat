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
	"proto/middleware"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RoomService struct {
	profileClient *grpc_client.ProfileClient
	rRepo         room_repo.RoomRepoInterface
	rMRepo        room_repo.RoomMemberRepoInterface
	db            *gorm.DB
	log           *logrus.Logger
}

func NewRoomService(profileClient *grpc_client.ProfileClient, rRepo room_repo.RoomRepoInterface,
	rMRepo room_repo.RoomMemberRepoInterface, db *gorm.DB, log *logrus.Logger) RoomServiceInterface {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &RoomService{
		profileClient: profileClient,
		rRepo:         rRepo,
		db:            db,
		rMRepo:        rMRepo,
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
		r.log.Warn("User_id is not a string")
		return 0, middleware.NewCustomError(http.StatusInternalServerError, "user_id is not a string", nil)
	}
	userIdInt, err := strconv.ParseInt(userIdStr, 10, 64)

	if strings.TrimSpace(name) == "" {
		r.log.WithFields(logrus.Fields{
			"name": name,
		}).Warn("Room name is empty")
		return 0, middleware.NewCustomError(http.StatusBadRequest, "room name cannot be empty", nil)
	}

	var roomId int64
	var transactionErr error

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		room := &models.Room{Name: name}
		if err := r.rRepo.Create(txCtx, room); err != nil {
			r.log.WithError(err).Warn("Failed to create room")
			transactionErr = fmt.Errorf("failed to create room: %w", err)
			return err
		}

		if err := r.rMRepo.AddMember(txCtx, room.Id, userIdInt); err != nil {
			r.log.WithError(err).Warn("Failed to add creator to room")
			transactionErr = fmt.Errorf("failed to add creator to room: %w", err)
			return err
		}

		if err := r.rMRepo.SetAdmin(txCtx, room.Id, userIdInt, true); err != nil {
			r.log.WithError(err).Warn("Failed to set admin status")
			transactionErr = fmt.Errorf("failed to set admin status: %w", err)
			return err
		}
		roomId = room.Id
		return nil
	})

	if err != nil {
		if transactionErr != nil {
			return 0, middleware.NewCustomError(http.StatusInternalServerError, transactionErr.Error(), transactionErr)
		}
		return 0, middleware.NewCustomError(http.StatusInternalServerError, "transaction failed", err)
	}

	r.log.WithFields(logrus.Fields{
		"room_id": roomId,
		"name":    name,
		"admin":   userIdInt,
	}).Info("Room created successfully")

	return roomId, nil
}

func (r *RoomService) RenameRoomById(ctx context.Context, roomId int64, name string) error {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		r.log.WithFields(logrus.Fields{
			"error": "context is not a gin context",
		}).Warn("Failed to get gin context")
		return middleware.NewCustomError(http.StatusInternalServerError, "context is not a gin context", errors.New("context is not a gin context"))
	}
	userId, exists := ginCtx.Get("user_id")
	if !exists {
		r.log.WithFields(logrus.Fields{
			"error": "user_id is not a string",
		}).Warn("Failed to convert user_id to string")
		return middleware.NewCustomError(http.StatusUnauthorized, "user_id not found in context", nil)
	}
	userIdStr, ok := userId.(string)
	if !ok {
		r.log.Warn("User_id is not a string")
		return middleware.NewCustomError(http.StatusInternalServerError, "user_id is not a string", nil)
	}

	userIdInt, err := strconv.ParseInt(userIdStr, 10, 64)

	if roomId <= 0 {
		r.log.Errorf("Room id %d is invalid", roomId)
		return middleware.NewCustomError(http.StatusBadRequest, "room id is invalid", nil)
	}

	if strings.TrimSpace(name) == "" {
		r.log.WithFields(logrus.Fields{
			"name": name,
		}).Warn("Room name is empty")
		return middleware.NewCustomError(http.StatusBadRequest, "room name cannot be empty", nil)
	}

	var transactionError error
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		roomMember, err := r.rMRepo.GetMemberByUserId(txCtx, roomId, userIdInt)
		if err != nil {
			r.log.WithFields(logrus.Fields{
				"user_id": userIdInt,
				"error":   err,
			}).Warn("Failed to get room member")
			transactionError = fmt.Errorf("failed to get room member: %w", err)
			return err
		}

		if roomMember == nil || !roomMember.IsAdmin {
			r.log.WithField("user_id", userIdInt).Warn("User is not admin")
			transactionError = errors.New("user is not admin")
			return transactionError
		}

		updateData := &models.Room{Name: name}
		if err := r.rRepo.Update(txCtx, roomId, updateData); err != nil {
			r.log.WithError(err).Warn("Failed to update room")
			transactionError = fmt.Errorf("failed to update room: %w", err)
			return err
		}

		return nil
	})

	if err != nil {
		if transactionError != nil {
			return middleware.NewCustomError(http.StatusInternalServerError, transactionError.Error(), transactionError)
		}
		return middleware.NewCustomError(http.StatusInternalServerError, "transaction failed", err)
	}

	r.log.WithFields(logrus.Fields{
		"room_id":  roomId,
		"new_name": name,
	}).Debug("Room renamed successfully")

	return nil
}

func (r *RoomService) DeleteRoomById(ctx context.Context, roomId int64) error {
	//TODO implement me
	panic("implement me")
}

func (r *RoomService) GetRoomById(ctx context.Context, roomId int64) (*models.Room, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RoomService) GetRoomList(ctx context.Context, search string, limit, offset int) ([]*models.Room, error) {
	//TODO implement me
	panic("implement me")
}
