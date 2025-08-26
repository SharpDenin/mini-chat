package room_repo

import (
	"chat_service/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RoomMemberRepo struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewRoomMemberRepo(db *gorm.DB, log *logrus.Logger) RoomMemberRepoInterface {
	return &RoomMemberRepo{
		db:  db,
		log: log,
	}
}

func (r *RoomMemberRepo) AddMember(ctx context.Context, roomId, userId int64) error {
	if userId <= 0 {
		r.log.WithFields(logrus.Fields{"userId": userId}).Error("Invalid user id")
		return fmt.Errorf("invalid user Id: %d", userId)
	}
	if roomId <= 0 {
		r.log.WithFields(logrus.Fields{"roomId": roomId}).Error("Invalid room id")
		return fmt.Errorf("invalid room id: %d", roomId)
	}

	tx := r.db.Begin().WithContext(ctx)
	commited := false
	defer func() {
		if !commited {
			tx.Rollback()
		}
	}()

	// Проверка на существование пользователя по UserId - gRPC на слое сервиса
	// Проверка на существование пользователя по RoomId - на слое сервиса

	var existingMember models.RoomMember
	if err := tx.Where("room_id = ? AND user_id = ?", roomId, userId).
		First(&existingMember).Error; err == nil {
		r.log.Error("User is already member of room")
		return fmt.Errorf("user is already member of room")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("error: %w", err)
	}

	member := models.RoomMember{
		RoomId: roomId,
		UserId: userId,
	}

	if err := tx.Create(&member).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": roomId}).Error("Failed to create room member")
		return fmt.Errorf("create room member: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": roomId}).Error("Failed to commit")
		return fmt.Errorf("commit room member: %w", err)
	}
	commited = true

	return nil
}

func (r *RoomMemberRepo) RemoveMember(ctx context.Context, roomId, userId int64) error {
	//TODO implement me
	panic("implement me")
}

func (r *RoomMemberRepo) GetMembersByRoom(ctx context.Context, roomId int64) ([]*models.RoomMember, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RoomMemberRepo) GetRoomsByUserId(ctx context.Context, userId int64) ([]*models.Room, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RoomMemberRepo) SetAdmin(ctx context.Context, roomMemberId int64, isAdmin bool) error {
	if roomMemberId <= 0 {
		r.log.WithFields(logrus.Fields{"roomMemberId": roomMemberId}).Error("Invalid room member id")
		return fmt.Errorf("invalid room member id: %d", roomMemberId)
	}
	tx := r.db.Begin().WithContext(ctx)
	commited := false
	defer func() {
		if !commited {
			tx.Rollback()
		}
	}()

	var existingMember models.RoomMember
	if err := tx.First(&existingMember, roomMemberId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("not found room member by id: %d", roomMemberId)
		}
		r.log.WithFields(logrus.Fields{"error": err, "id": roomMemberId}).Error("Failed to find room member")
		return fmt.Errorf("failed to find room member by id: %w", err)
	}
	if existingMember.IsAdmin == isAdmin {
		r.log.Errorf("User admin status already: %v", existingMember.IsAdmin)
		return fmt.Errorf("user admin status already: %v", existingMember.IsAdmin)
	}
	if err := tx.Model(&existingMember).Update("is_admin", isAdmin).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": roomMemberId}).Error("Failed to update")
		return fmt.Errorf("failed to update room member: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": roomMemberId}).Error("Failed to commit")
		return fmt.Errorf("commit room member: %w", err)
	}
	commited = true

	return nil
}
