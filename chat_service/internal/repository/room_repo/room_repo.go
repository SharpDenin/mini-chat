package room_repo

import (
	"chat_service/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RoomRepo struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewRoomRepo(db *gorm.DB, log *logrus.Logger) RoomRepoInterface {
	return &RoomRepo{
		db:  db,
		log: log,
	}
}

func (r *RoomRepo) Create(ctx context.Context, room *models.Room) error {
	if err := r.db.WithContext(ctx).
		Create(room).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err}).Error("Failed to create room")
		return fmt.Errorf("create room error: %w", err)
	}

	return nil
}

func (r *RoomRepo) GetRoomById(ctx context.Context, id int64) (*models.Room, error) {
	var room models.Room
	err := r.db.WithContext(ctx).First(&room, id).Error
	if err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to get room by Id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("room not found: %w", err)
		}
		return nil, fmt.Errorf("get room by id error: %w", err)
	}

	return &room, nil
}

func (r *RoomRepo) GetAll(ctx context.Context, searchFilter string, limit, offset int) ([]*models.Room, error) {
	query := r.db.WithContext(ctx).Model(&models.Room{})
	if searchFilter != "" {
		query = query.Where("name LIKE ?", "%"+searchFilter+"%")
	}

	var rooms []*models.Room
	if err := query.
		Limit(limit).
		Offset(offset).
		Find(&rooms).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err}).Error("Failed to get rooms list")
		return nil, fmt.Errorf("get rooms error: %w", err)
	}

	return rooms, nil
}

func (r *RoomRepo) Update(ctx context.Context, id int64, room *models.Room) error {
	var existingRoom models.Room
	if err := r.db.WithContext(ctx).
		First(&existingRoom, id).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to get room by Id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("room not found: %w", err)
		}
		return fmt.Errorf("get room by id error: %w", err)
	}

	updates := map[string]interface{}{}
	if room.Name != "" {
		updates["name"] = room.Name
	}

	if err := r.db.WithContext(ctx).
		Model(&existingRoom).Updates(updates).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to update room by Id")
		return fmt.Errorf("update room by id error: %w", err)
	}

	if err := r.db.WithContext(ctx).
		First(&existingRoom, id).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to fetch updated room")
		return fmt.Errorf("failed to fetch updated room: %w", err)
	}

	return nil
}

func (r *RoomRepo) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).
		Delete(&models.Room{}, id).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to delete room by Id")
		return fmt.Errorf("delete room by id error: %w", err)
	}

	return nil
}
