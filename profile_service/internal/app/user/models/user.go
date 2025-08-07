package models

import (
	aModel "profile_service/internal/app/auth/model"
	"time"

	"gorm.io/gorm"
)

type User struct {
	Id        int64          `gorm:"primaryKey;autoIncrement;column:id"`
	Username  string         `gorm:"type:text;not null;uniqueIndex:username_unique"`
	Email     string         `gorm:"type:text;not null;uniqueIndex:email_unique"`
	Password  string         `gorm:"type:text;not null"`
	CreatedAt time.Time      `gorm:"type:timestamp with time zone;default:now()"`
	UpdatedAt time.Time      `gorm:"type:timestamp with time zone;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamp with time zone;index"`

	AuthTokens []aModel.AuthToken `gorm:"foreignKey:UserId;constraint:OnDelete:CASCADE"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now().UTC()
	return nil
}
