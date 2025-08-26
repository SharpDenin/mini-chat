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
	if room == nil {
		r.log.Error("Create room error: room is nil")
		return fmt.Errorf("create room error: room is nil")
	}

	tx := r.db.Begin().WithContext(ctx)
	commited := false
	defer func() {
		if !commited {
			tx.Rollback()
		}
	}()

	if err := tx.Create(room).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err}).Error("Failed to create room")
		return fmt.Errorf("create room error: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err}).Error("Failed to to commit transaction")
		return fmt.Errorf("failed to commit: %w", err)
	}
	commited = true

	return nil
}

func (r *RoomRepo) GetById(ctx context.Context, id int64) (*models.Room, error) {
	if id <= 0 {
		r.log.WithFields(logrus.Fields{"id": id}).Error("Invalid room id")
		return nil, fmt.Errorf("invalid user ID: %d", id)
	}

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

func (r *RoomRepo) GetAll(ctx context.Context, searchFilter string) ([]*models.Room, error) {
	query := r.db.WithContext(ctx).Model(&models.Room{})

	if searchFilter != "" {
		query = query.Where("name LIKE ?", "%"+searchFilter+"%")
	}

	var rooms []*models.Room

	if err := query.
		Find(&rooms).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err}).Error("Failed to get rooms list")
		return nil, fmt.Errorf("get rooms error: %w", err)
	}

	return rooms, nil
}

func (r *RoomRepo) Update(ctx context.Context, id int64, room *models.Room) error {
	if id <= 0 {
		r.log.WithFields(logrus.Fields{"id": id}).Error("Invalid room id")
		return fmt.Errorf("invalid user ID: %d", id)
	}
	if room == nil {
		r.log.WithFields(logrus.Fields{"id": id}).Warn("Room is nil")
		return nil
	}

	tx := r.db.Begin().WithContext(ctx)
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	var existingRoom models.Room

	if err := tx.First(&existingRoom, id).Error; err != nil {
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

	if err := tx.Model(&existingRoom).Updates(updates).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to update room by Id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("room not found: %w", err)
		}
		return fmt.Errorf("update room by id error: %w", err)
	}

	if err := tx.First(&existingRoom, id).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to fetch updated room")
		return fmt.Errorf("failed to fetch updated room: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to commit transaction")
		return fmt.Errorf("failed to commit: %w", err)
	}
	committed = true

	return nil
}

func (r *RoomRepo) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		r.log.WithFields(logrus.Fields{"id": id}).Error("Invalid room id")
		return fmt.Errorf("invalid user ID: %d", id)
	}

	tx := r.db.Begin().WithContext(ctx)
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	if err := tx.Delete(&models.Room{}, id).Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to delete room by Id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("room not found: %w", err)
		}
		return fmt.Errorf("delete room by id error: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		r.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to commit transaction")
		return fmt.Errorf("failed to commit: %w", err)
	}
	committed = true

	return nil
}
