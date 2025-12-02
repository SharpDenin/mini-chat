package models

import (
	"time"

	"gorm.io/gorm"
)

type Room struct {
	Id        int64          `gorm:"primaryKey;autoIncrement;column:id"`
	Name      string         `gorm:"type:varchar(255);not_null;column:name"`
	CreatedAt time.Time      `gorm:"type:timestamp with time zone;default:now()"`
	UpdatedAt time.Time      `gorm:"type:timestamp with time zone;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamp with time zone;index"`
}

func (r *Room) BeforeCreate(tx *gorm.DB) error {
	r.CreatedAt = time.Now().UTC()
	r.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *Room) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now().UTC()
	return nil
}
