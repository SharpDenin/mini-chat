package models

import (
	"gorm.io/gorm"
	"time"
)

type BlockedUser struct {
	Id        int64          `gorm:"primaryKey;autoIncrement;column:id"`
	BlockerId int64          `gorm:"column:blocker_id;not null;index:idx_blocker;index:idx_block_pair,unique"`
	BlockedId int64          `gorm:"column:blocked_id;not null;index:idx_blocked;index:idx_block_pair,unique"`
	Reason    string         `gorm:"column:reason;type:text"`
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamp with time zone;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp with time zone;index"`
}

func (bu *BlockedUser) BeforeCreate(tx *gorm.DB) error {
	bu.CreatedAt = time.Now().UTC()
	return nil
}
