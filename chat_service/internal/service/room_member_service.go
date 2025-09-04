package service

import (
	"chat_service/internal/models"
	"chat_service/internal/repository/room_repo"
	"chat_service/internal/service/helper"
	"chat_service/pkg/grpc_client"
	"context"
	"errors"
	"net/http"
	"os"
	"proto/generated/profile"
	"proto/middleware"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RoomMemberService struct {
	profileClient *grpc_client.ProfileClient
	rRepo         room_repo.RoomRepoInterface
	rMRepo        room_repo.RoomMemberRepoInterface
	db            *gorm.DB
	log           *logrus.Logger
}

func NewRoomMemberService(profileClient *grpc_client.ProfileClient, rRepo room_repo.RoomRepoInterface,
	rMRepo room_repo.RoomMemberRepoInterface, db *gorm.DB, log *logrus.Logger) RoomMemberServiceInterface {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &RoomMemberService{
		profileClient: profileClient,
		rRepo:         rRepo,
		rMRepo:        rMRepo,
		db:            db,
		log:           log,
	}
}

func (r *RoomMemberService) AddMember(ctx context.Context, roomId, userId int64) error {
	if roomId <= 0 {
		r.log.Errorf("Room id %d is invalid", roomId)
		return middleware.NewCustomError(http.StatusBadRequest, "room id is invalid", nil)
	}
	if userId <= 0 {
		r.log.Errorf("Room id %d is invalid", userId)
		return middleware.NewCustomError(http.StatusBadRequest, "room id is invalid", nil)
	}

	currentUserId, err := helper.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}
	if err = r.validateUserIsAdmin(ctx, roomId, currentUserId); err != nil {
		return middleware.NewCustomError(http.StatusForbidden, err.Error(), err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	userReq := &profile.UserExistsRequest{UserId: strconv.FormatInt(userId, 10)}
	exist, err := helper.CheckUserExist(ctx, r.profileClient, userReq)
	if err != nil {
		r.log.Error("Failed to check user existence", "error", err)
		return middleware.NewCustomError(http.StatusNotFound, "failed to verify user", err)
	}
	if !exist {
		return middleware.NewCustomError(http.StatusNotFound, "user does not exist", nil)
	}

	_, err = r.rRepo.GetById(ctx, roomId)
	if err != nil {
		if strings.Contains(err.Error(), "room not found") {
			r.log.WithFields(logrus.Fields{"room_id": roomId}).Warn("Room does not exist")
			return middleware.NewCustomError(http.StatusNotFound, "Room does not exist", nil)
		}
		r.log.WithFields(logrus.Fields{
			"room_id": roomId,
			"error":   err,
		}).Error("Failed to get room by Id")
		return middleware.NewCustomError(http.StatusInternalServerError, "Failed to verify room existence", err)
	}

	if err = r.rMRepo.AddMember(ctx, roomId, userId); err != nil {
		r.log.WithFields(logrus.Fields{
			"room_id": roomId,
			"error":   err,
		}).Error("Failed to add member to room")
		return middleware.NewCustomError(http.StatusInternalServerError, "Failed to add member to room", err)
	}

	r.log.WithFields(logrus.Fields{
		"room_id": roomId,
		"user_id": userId,
	}).Info("Member added successfully")

	return nil
}

func (r *RoomMemberService) RemoveMember(ctx context.Context, roomId, userId int64) error {
	//TODO implement me
	panic("implement me")
}

func (r *RoomMemberService) ListMembers(ctx context.Context, roomId int64) ([]*models.RoomMember, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RoomMemberService) SetAdmin(ctx context.Context, roomId, userId int64, isAdmin bool) error {
	//TODO implement me
	panic("implement me")
}

func (r *RoomMemberService) ListUserRooms(ctx context.Context, userId int64) ([]*models.Room, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RoomMemberService) validateUserIsAdmin(ctx context.Context, roomId, userId int64) error {
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
