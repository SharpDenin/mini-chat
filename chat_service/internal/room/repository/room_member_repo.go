package repository

import (
	"chat_service/internal/room/models"
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
	var existingMember models.RoomMember
	if err := r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomId, userId).
		First(&existingMember).Error; err == nil {
		r.log.Error("User is already member of room")
		return fmt.Errorf("user is already member of room")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	member := models.RoomMember{
		RoomId: roomId,
		UserId: userId,
	}

	if err := r.db.WithContext(ctx).
		Create(&member).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": roomId}).Error("Failed to create room member")
		return fmt.Errorf("create room member: %w", err)
	}

	return nil
}

func (r *RoomMemberRepo) GetMemberByUserId(ctx context.Context, roomId, userId int64) (*models.RoomMember, error) {
	var member models.RoomMember
	if err := r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomId, userId).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Warnf("room member not found (roomId=%d, userId=%d)", roomId, userId)
			return nil, nil
		}
		r.log.WithFields(logrus.Fields{"error": err, "id": roomId}).Error("failed to get room member")
		return nil, fmt.Errorf("failed to get room member: %w", err)
	}
	r.log.Debugf("member found (roomId=%d, userId=%d)", roomId, userId)
	return &member, nil
}

func (r *RoomMemberRepo) RemoveMember(ctx context.Context, roomId, userId int64) error {
	var existingMember models.RoomMember
	if err := r.db.WithContext(ctx).
		First(&existingMember, "room_id = ? AND user_id = ?", roomId, userId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Errorf("User: %d not found in room: %d", userId, roomId)
			return nil
		}
		r.log.WithFields(logrus.Fields{"error": err, "id": roomId}).Error("Failed to find user")
		return fmt.Errorf("find user: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Delete(&existingMember).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": roomId}).Error("Failed to remove user")
		return fmt.Errorf("failed to remove user: %w", err)
	}

	return nil
}

func (r *RoomMemberRepo) RemoveAllMembers(ctx context.Context, roomId int64) error {
	result := r.db.WithContext(ctx).
		Where("room_id = ?", roomId).
		Delete(&models.RoomMember{})

	if result.Error != nil {
		r.log.WithError(result.Error).
			WithField("room_id", roomId).
			Error("Failed to delete all room members")
		return fmt.Errorf("failed to delete all members: %w", result.Error)
	}

	r.log.WithFields(logrus.Fields{
		"room_id":       roomId,
		"rows_affected": result.RowsAffected,
	}).Debug("Deleted all room members")

	return nil
}

func (r *RoomMemberRepo) GetMembersByRoom(ctx context.Context, roomId int64) ([]*models.RoomMember, error) {
	var members []*models.RoomMember
	if err := r.db.WithContext(ctx).Where("room_id = ?", roomId).Find(&members).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": roomId}).Error("Failed to find user")
		return nil, fmt.Errorf("find user: %w", err)
	}

	return members, nil
}

func (r *RoomMemberRepo) GetRoomsByUserId(ctx context.Context, userId int64) ([]*models.Room, error) {
	var rooms []*models.Room
	err := r.db.WithContext(ctx).
		Table("rooms").
		Joins("JOIN room_members ON room_members.room_id = rooms.id").
		Where("room_members.user_id = ?", userId).
		Find(&rooms).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "user_id": userId}).Error("Failed to get rooms by user")
		return nil, fmt.Errorf("get rooms by user error: %w", err)
	}

	return rooms, nil
}

func (r *RoomMemberRepo) SetAdmin(ctx context.Context, roomId, userId int64, isAdmin bool) error {
	var existingMember models.RoomMember
	if err := r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomId, userId).First(&existingMember).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get room member: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Model(&existingMember).Update("is_admin", isAdmin).Error; err != nil {
		return fmt.Errorf("failed to update is_admin: %w", err)
	}

	return nil
}
