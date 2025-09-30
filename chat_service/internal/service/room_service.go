package service

import (
	"chat_service/internal/models"
	"chat_service/internal/repository/room_repo"
	"chat_service/internal/service/dto"
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

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		type txKey struct{}
		txCtx := context.WithValue(ctx, txKey{}, tx)

		room := &models.Room{Name: name}
		if err := r.rRepo.Create(txCtx, room); err != nil {
			r.log.WithError(err).Warn("Failed to create room")
			return fmt.Errorf("failed to create room: %w", err)
		}

		if err := r.rMRepo.AddMember(txCtx, room.Id, userIdInt); err != nil {
			r.log.WithError(err).Warn("Failed to add creator to room")
			return fmt.Errorf("failed to add creator to room: %w", err)
		}

		if err := r.rMRepo.SetAdmin(txCtx, room.Id, userIdInt, true); err != nil {
			r.log.WithError(err).Warn("Failed to set admin status")
			return fmt.Errorf("failed to set admin status: %w", err)
		}
		roomId = room.Id
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("room creation transaction failed: %w", err)
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

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		type txKey struct{}
		txCtx := context.WithValue(ctx, txKey{}, tx)

		room, err := r.rRepo.GetRoomById(txCtx, roomId)
		if err != nil {
			r.log.WithError(err).Warn("Failed to get room")
			return fmt.Errorf("failed to get room: %w", err)
		}
		if room == nil {
			r.log.WithField("room_id", roomId).Warn("Room not found")
			return errors.New("room not found")
		}

		if err = r.rMRepo.RemoveAllMembers(txCtx, roomId); err != nil {
			r.log.WithError(err).Warn("Failed to remove member")
			return fmt.Errorf("failed to remove room members: %w", err)
		}

		if err = r.rRepo.Delete(txCtx, roomId); err != nil {
			r.log.WithError(err).Warn("Failed to delete room")
			return fmt.Errorf("failed to delete room: %w", err)
		}

		return nil
	})
	if err != nil {
		return middleware.NewCustomError(http.StatusInternalServerError, "transaction failed", err)
	}

	r.log.WithFields(logrus.Fields{
		"room_id": roomId,
	}).Debug("Room deleted successfully")

	return nil
}

func (r *RoomService) GetRoomById(ctx context.Context, roomId int64) (*dto.GetRoomResponse, error) {
	if roomId <= 0 {
		r.log.Errorf("Room id %d is invalid", roomId)
		return nil, middleware.NewCustomError(http.StatusBadRequest, "room id is invalid", nil)
	}

	room, err := r.rRepo.GetRoomById(ctx, roomId)
	if err != nil {
		r.log.WithError(err).Warn("Failed to get room")
		return nil, err
	}
	return &dto.GetRoomResponse{
		Id:   room.Id,
		Name: room.Name,
	}, nil
}

func (r *RoomService) GetRoomList(ctx context.Context, filter *dto.SearchFilter) ([]*dto.GetRoomResponse, error) {
	if len(filter.Search) > 100 {
		filter.Search = filter.Search[:100]
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	roomList, err := r.rRepo.GetAll(ctx, filter.Search, filter.Limit, filter.Offset)
	if err != nil {
		r.log.WithError(err).Warn("Failed to get room list")
		return nil, err
	}

	resp := make([]*dto.GetRoomResponse, len(roomList))
	for i, room := range roomList {
		resp[i] = &dto.GetRoomResponse{
			Id:   room.Id,
			Name: room.Name,
		}
	}

	return resp, nil
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
