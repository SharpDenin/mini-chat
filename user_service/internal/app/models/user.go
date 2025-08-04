package models

import (
	"gorm.io/gorm"
	"time"
	aModel "user_service/internal/app/auth/model"
)

type User struct {
	gorm.Model
	Id        int64     `json:"id" gorm:"column:id" validate:"required"`
	Username  string    `json:"username" gorm:"column:username" validate:"required, min=1, max=50"`
	Email     string    `json:"email" gorm:"column:email" validate:"required,email"`
	Password  string    `json:"password" gorm:"column:password;size:255" validate:"required"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`

	AuthTokens []aModel.AuthToken `gorm:"foreignKey:UserId;constraint:OnDelete:CASCADE"`
}
