package models

import (
	"gorm.io/gorm"
	"time"
)

// FriendRequest - модель запроса в друзья
type FriendRequest struct {
	Id         int64  `gorm:"primaryKey;autoIncrement;column:id"`
	SenderId   int64  `gorm:"column:sender_id;not null;index:idx_request_sender;index:idx_request_pair,unique"`
	ReceiverId int64  `gorm:"column:receiver_id;not null;index:idx_request_receiver;index:idx_request_pair,unique"`
	Status     string `gorm:"column:status;type:varchar(20);not null;default:'pending'"` // pending, accepted, rejected, cancelled
	Message    string `gorm:"column:message;type:text"`
	//ExpiresAt  *time.Time     `gorm:"column:expires_at;type:timestamp with time zone"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp with time zone;default:now()"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp with time zone;default:now()"`
	//DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp with time zone;index"`
}

func (fr *FriendRequest) BeforeCreate(tx *gorm.DB) error {
	fr.CreatedAt = time.Now().UTC()
	fr.UpdatedAt = time.Now().UTC()
	return nil
}

func (fr *FriendRequest) BeforeUpdate(tx *gorm.DB) error {
	fr.UpdatedAt = time.Now().UTC()
	return nil
}
