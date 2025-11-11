package service

import (
	"chat_service/internal/room/repository"
	"chat_service/internal/room/service/dto"
	"chat_service/internal/room/service/helper"
	"chat_service/middleware_chat"
	"chat_service/pkg/grpc_client"
	"chat_service/pkg/grpc_generated/profile"
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RoomMemberService struct {
	profileClient *grpc_client.ProfileClient
	rRepo         repository.RoomRepoInterface
	rMRepo        repository.RoomMemberRepoInterface
	db            *gorm.DB
	log           *logrus.Logger
}

func NewRoomMemberService(profileClient *grpc_client.ProfileClient, rRepo repository.RoomRepoInterface,
	rMRepo repository.RoomMemberRepoInterface, db *gorm.DB, log *logrus.Logger) RoomMemberServiceInterface {
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
	if err := r.validateBaseParams(roomId, userId); err != nil {
		return err
	}

	currentUserId, err := helper.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware_chat.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}
	if err = r.validateUserIsAdmin(ctx, roomId, currentUserId); err != nil {
		return middleware_chat.NewCustomError(http.StatusForbidden, err.Error(), err)
	}

	if err := r.validateUserExists(ctx, userId); err != nil {
		return middleware_chat.NewCustomError(http.StatusNotFound, err.Error(), nil)
	}

	if err := r.validateRoomExists(ctx, roomId); err != nil {
		return middleware_chat.NewCustomError(http.StatusNotFound, err.Error(), nil)
	}

	member, err := r.rMRepo.GetMemberByUserId(ctx, roomId, userId)
	if err != nil {
		r.log.WithError(err).Warn("Failed to check existing membership")
	} else if member != nil {
		r.log.WithField("user_id", userId).Info("User already exists in room")
		return middleware_chat.NewCustomError(http.StatusConflict, "user already exists in room", nil)
	}

	if err = r.rMRepo.AddMember(ctx, roomId, userId); err != nil {
		r.log.WithFields(logrus.Fields{
			"room_id": roomId,
			"error":   err,
		}).Error("Failed to add member to room")
		return middleware_chat.NewCustomError(http.StatusInternalServerError, "failed to add member to room", err)
	}

	r.log.WithFields(logrus.Fields{
		"room_id": roomId,
		"user_id": userId,
	}).Info("Member added successfully")

	return nil
}

func (r *RoomMemberService) RemoveMember(ctx context.Context, roomId, userId int64) error {
	if err := r.validateBaseParams(roomId, userId); err != nil {
		return err
	}

	if err := r.validateUserExists(ctx, userId); err != nil {
		return middleware_chat.NewCustomError(http.StatusNotFound, err.Error(), nil)
	}

	if err := r.validateRoomExists(ctx, roomId); err != nil {
		return middleware_chat.NewCustomError(http.StatusNotFound, err.Error(), nil)
	}

	currentUserId, err := helper.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware_chat.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}
	if err = r.validateUserIsAdmin(ctx, roomId, currentUserId); err != nil {
		return middleware_chat.NewCustomError(http.StatusForbidden, err.Error(), err)
	}

	member, err := r.rMRepo.GetMemberByUserId(ctx, roomId, userId)
	if err != nil {
		r.log.WithError(err).Error("Failed to check membership")
		return middleware_chat.NewCustomError(http.StatusInternalServerError, "failed to verify membership", err)
	}
	if member == nil {
		r.log.WithFields(logrus.Fields{
			"room_id": roomId,
			"user_id": userId,
		}).Info("User not found in room")
		return middleware_chat.NewCustomError(http.StatusNotFound, "user not in room", nil)
	}
	if member.IsAdmin {
		r.log.WithFields(logrus.Fields{"user_id": userId}).Info("User is admin")
		return middleware_chat.NewCustomError(http.StatusForbidden, "user is admin", nil)
	}

	if err = r.rMRepo.RemoveMember(ctx, roomId, userId); err != nil {
		r.log.WithFields(logrus.Fields{
			"room_id": roomId,
			"error":   err,
		}).Error("Failed to remove member from room")
		return middleware_chat.NewCustomError(http.StatusInternalServerError, "failed to remove member from room", err)
	}

	r.log.WithFields(logrus.Fields{
		"room_id": roomId,
		"user_id": userId,
	}).Info("Member removed successfully")

	return nil
}

func (r *RoomMemberService) ListMembers(ctx context.Context, roomId int64) ([]*dto.GetRoomMemberResponse, error) {
	if roomId <= 0 {
		r.log.Errorf("Room id %d is invalid", roomId)
		return nil, middleware_chat.NewCustomError(http.StatusBadRequest, "room id is invalid", nil)
	}

	if err := r.validateRoomExists(ctx, roomId); err != nil {
		return nil, middleware_chat.NewCustomError(http.StatusNotFound, err.Error(), nil)
	}

	members, err := r.rMRepo.GetMembersByRoom(ctx, roomId)
	if err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "room_id": roomId}).Error("Failed to get room members")
		return nil, middleware_chat.NewCustomError(http.StatusInternalServerError, "failed to get room members", err)
	}

	resp := make([]*dto.GetRoomMemberResponse, len(members))
	for i, member := range members {
		resp[i] = &dto.GetRoomMemberResponse{
			UserId:  member.UserId,
			RoomId:  member.RoomId,
			IsAdmin: member.IsAdmin,
		}
	}

	return resp, nil
}

func (r *RoomMemberService) SetAdmin(ctx context.Context, roomId, userId int64, isAdmin bool) error {
	if err := r.validateBaseParams(roomId, userId); err != nil {
		return err
	}

	currentUserId, err := helper.GetUserIdFromContext(ctx)
	if err != nil {
		return middleware_chat.NewCustomError(http.StatusUnauthorized, err.Error(), nil)
	}
	if err = r.validateUserIsAdmin(ctx, roomId, currentUserId); err != nil {
		return middleware_chat.NewCustomError(http.StatusForbidden, err.Error(), err)
	}

	if err := r.validateUserExists(ctx, userId); err != nil {
		return middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	if err := r.validateRoomExists(ctx, roomId); err != nil {
		return middleware_chat.NewCustomError(http.StatusNotFound, err.Error(), nil)
	}

	member, err := r.rMRepo.GetMemberByUserId(ctx, roomId, userId)
	if err != nil {
		r.log.WithError(err).Error("Failed to check membership")
		return middleware_chat.NewCustomError(http.StatusInternalServerError, "failed to verify membership", err)
	}
	if member == nil {
		r.log.WithFields(logrus.Fields{
			"room_id": roomId,
			"user_id": userId,
		}).Info("User not found in room")
		return middleware_chat.NewCustomError(http.StatusNotFound, "user not in room", nil)
	}

	if member.IsAdmin == isAdmin {
		r.log.Infof("User already has admin status: %t", isAdmin)
		return nil
	}

	if err = r.rMRepo.SetAdmin(ctx, roomId, userId, isAdmin); err != nil {
		return middleware_chat.NewCustomError(http.StatusInternalServerError, "failed to set admin", err)
	}

	r.log.Info("Admin status updated successfully")

	return nil
}

func (r *RoomMemberService) ListUserRooms(ctx context.Context, userId int64) ([]*dto.GetRoomResponse, error) {
	if userId <= 0 {
		r.log.Errorf("Room id %d is invalid", userId)
		return nil, middleware_chat.NewCustomError(http.StatusBadRequest, "room id is invalid", nil)
	}

	if err := r.validateUserExists(ctx, userId); err != nil {
		return nil, err
	}

	rooms, err := r.rMRepo.GetRoomsByUserId(ctx, userId)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"user_id": userId,
			"error":   err,
		}).Error("Failed to get room by UserId")
		return nil, middleware_chat.NewCustomError(http.StatusInternalServerError, "failed to get room by UserId", err)
	}

	resp := make([]*dto.GetRoomResponse, len(rooms))
	for i, room := range rooms {
		resp[i] = &dto.GetRoomResponse{
			Id:   room.Id,
			Name: room.Name,
		}
	}

	r.log.WithFields(logrus.Fields{
		"user_id": userId,
		"rooms":   len(rooms),
	})

	return resp, nil
}

func (r *RoomMemberService) validateBaseParams(roomId, userId int64) error {
	if roomId <= 0 {
		return middleware_chat.NewCustomError(http.StatusBadRequest, "room id is invalid", nil)
	}
	if userId <= 0 {
		return middleware_chat.NewCustomError(http.StatusBadRequest, "user id is invalid", nil)
	}
	return nil
}

func (r *RoomMemberService) validateUserExists(ctx context.Context, userId int64) error {
	userReq := &profile.UserExistsRequest{UserId: strconv.FormatInt(userId, 10)}
	exist, err := helper.CheckUserExist(ctx, r.profileClient, userReq)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"user_id": userId,
			"error":   err,
		}).Warn("Failed to get user")
		return errors.New("failed to get user")
	}
	if !exist {
		r.log.WithField("user_id", userId).Warn("User not found")
		return errors.New("user not found")
	}
	return nil
}

func (r *RoomMemberService) validateRoomExists(ctx context.Context, roomId int64) error {
	room, err := r.rRepo.GetRoomById(ctx, roomId)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"room_id": roomId,
			"error":   err,
		}).Warn("Failed to get room")
		return errors.New("failed to get room")
	}
	if room == nil {
		r.log.WithField("room_id", roomId).Warn("Room not found")
		return errors.New("room not found")
	}

	return nil
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
	if roomMember == nil {
		r.log.WithFields(logrus.Fields{
			"user_id": userId,
			"error":   err,
		}).Warn("Failed to get room member")
		return errors.New("user is not member of the room")
	}
	if !roomMember.IsAdmin {
		r.log.WithField("user_id", userId).Warn("User is not admin")
		return errors.New("user is not admin")
	}

	return nil
}
