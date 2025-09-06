package service

import (
	"chat_service/internal/models"
	"chat_service/internal/repository/room_repo"
	"chat_service/internal/service/helper"
	"chat_service/pkg/grpc_client"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"proto/middleware"
	"strings"

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
	userIdInt, err := helper.GetUserIdFromContext(ctx)
	if err != nil {
		return 0, middleware.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}

	if strings.TrimSpace(name) == "" {
		r.log.WithFields(logrus.Fields{
			"name": name,
		}).Warn("Room name is empty")
		return 0, middleware.NewCustomError(http.StatusBadRequest, "room name cannot be empty", nil)
	}

	var roomId int64
	var transactionErr error

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		type txKey struct{}
		txCtx := context.WithValue(ctx, txKey{}, tx)

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
	if roomId <= 0 {
		r.log.Errorf("Room id %d is invalid", roomId)
		return middleware.NewCustomError(http.StatusBadRequest, "room id is invalid", nil)
	}

	userId, err := helper.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}
	if err = r.validateUserIsAdmin(ctx, roomId, userId); err != nil {
		return middleware.NewCustomError(http.StatusForbidden, err.Error(), err)
	}

	if strings.TrimSpace(name) == "" {
		r.log.WithFields(logrus.Fields{
			"name": name,
		}).Warn("Room name is empty")
		return middleware.NewCustomError(http.StatusBadRequest, "room name cannot be empty", nil)
	}

	updateData := &models.Room{Name: name}
	if err := r.rRepo.Update(ctx, roomId, updateData); err != nil {
		r.log.WithError(err).Warn("Failed to update room")
		return middleware.NewCustomError(http.StatusInternalServerError, "failed to update room", err)
	}

	r.log.WithFields(logrus.Fields{
		"room_id":  roomId,
		"new_name": name,
	}).Debug("Room renamed successfully")

	return nil
}

func (r *RoomService) DeleteRoomById(ctx context.Context, roomId int64) error {
	if roomId <= 0 {
		r.log.Errorf("Room id %d is invalid", roomId)
		return middleware.NewCustomError(http.StatusBadRequest, "room id is invalid", nil)
	}

	userId, err := helper.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}
	if err = r.validateUserIsAdmin(ctx, roomId, userId); err != nil {
		return middleware.NewCustomError(http.StatusForbidden, err.Error(), err)
	}

	var transactionError error
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		type txKey struct{}
		txCtx := context.WithValue(ctx, txKey{}, tx)

		room, err := r.rRepo.GetRoomById(txCtx, roomId)
		if err != nil {
			r.log.WithError(err).Warn("Failed to get room")
			transactionError = fmt.Errorf("failed to get room: %w", err)
			return err
		}
		if room == nil {
			r.log.WithField("room_id", roomId).Warn("Room not found")
			transactionError = errors.New("room not found")
			return err
		}

		if err = r.rMRepo.RemoveAllMembers(txCtx, roomId); err != nil {
			transactionError = fmt.Errorf("failed to remove room members: %w", err)
			return err
		}

		if err = r.rRepo.Delete(txCtx, roomId); err != nil {
			r.log.WithError(err).Warn("Failed to delete room")
			transactionError = fmt.Errorf("failed to delete room: %w", err)
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
		"room_id": roomId,
	}).Debug("Room deleted successfully")

	return nil
}

func (r *RoomService) GetRoomById(ctx context.Context, roomId int64) (*models.Room, error) {
	if roomId <= 0 {
		r.log.Errorf("Room id %d is invalid", roomId)
		return nil, middleware.NewCustomError(http.StatusBadRequest, "room id is invalid", nil)
	}

	room, err := r.rRepo.GetRoomById(ctx, roomId)
	if err != nil {
		r.log.WithError(err).Warn("Failed to get room")
		return nil, err
	}
	return room, nil
}

func (r *RoomService) GetRoomList(ctx context.Context, search string, limit, offset int) ([]*models.Room, error) {
	if len(search) > 100 {
		search = search[:100]
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if limit > 100 {
		limit = 100
	}

	roomList, err := r.rRepo.GetAll(ctx, search, limit, offset)
	if err != nil {
		r.log.WithError(err).Warn("Failed to get room list")
		return nil, err
	}
	return roomList, nil
}

func (r *RoomService) validateUserIsAdmin(ctx context.Context, roomId, userId int64) error {
	roomMember, err := r.rMRepo.GetMemberByUserId(ctx, roomId, userId)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"user_id": userId,
			"error":   err,
		}).Warn("Failed to get room member")
		return errors.New("failed to get room member")
	}
	if roomMember == nil || !roomMember.IsAdmin {
		r.log.WithField("user_id", userId).Warn("User is not admin")
		return errors.New("user is not admin")
	}

	return nil
}
