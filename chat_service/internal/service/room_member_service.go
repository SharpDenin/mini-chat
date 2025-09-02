package service

import (
	"chat_service/internal/models"
	"chat_service/internal/repository/room_repo"
	"chat_service/pkg/grpc_client"
	"context"
	"fmt"
	"net/http"
	"os"
	"proto/generated/profile"
	"proto/middleware"
	"strconv"
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

func (r RoomMemberService) AddMember(ctx context.Context, roomId, userId int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	userReq := &profile.UserExistsRequest{UserId: strconv.FormatInt(userId, 10)}
	userResp, err := r.profileClient.UserExists(ctx, userReq)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"user_id": userId,
			"error":   err,
		}).Warn("Failed to check user existence")
		return middleware.NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to check user existence: %v", err), err)
	}
	if !userResp.Exists {
		r.log.WithFields(logrus.Fields{
			"user_id": userId,
		}).Warn("User does not exist")
		return middleware.NewCustomError(http.StatusNotFound, "user does not exist", nil)
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return nil
	})
	if err != nil {
		return middleware.NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to add member to room: %v", err), err)
	}
	//TODO implement me
	panic("implement me")
}

func (r RoomMemberService) RemoveMember(ctx context.Context, roomId, userId int64) error {
	//TODO implement me
	panic("implement me")
}

func (r RoomMemberService) ListMembers(ctx context.Context, roomId int64) ([]*models.RoomMember, error) {
	//TODO implement me
	panic("implement me")
}

func (r RoomMemberService) SetAdmin(ctx context.Context, roomId, userId int64, isAdmin bool) error {
	//TODO implement me
	panic("implement me")
}

func (r RoomMemberService) ListUserRooms(ctx context.Context, userId int64) ([]*models.Room, error) {
	//TODO implement me
	panic("implement me")
}
