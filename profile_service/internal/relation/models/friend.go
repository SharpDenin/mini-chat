package models

import (
	"gorm.io/gorm"
	"time"
)

// Friend - модель установленной дружбы
type Friend struct {
	Id        int64          `gorm:"primaryKey;autoIncrement;column:id"`
	UserId    int64          `gorm:"column:user_id;not null;index:idx_friend_user;index:idx_friend_pair,unique"`
	FriendId  int64          `gorm:"column:friend_id;not null;index:idx_friend_friend;index:idx_friend_pair,unique"`
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamp with time zone;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp with time zone;index"`
}

func (f *Friend) BeforeCreate(tx *gorm.DB) error {
	if f.UserId > f.FriendId {
		f.UserId, f.FriendId = f.FriendId, f.UserId
	}
	f.CreatedAt = time.Now().UTC()
	return nil
}
