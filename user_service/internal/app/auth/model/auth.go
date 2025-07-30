package model

import "time"

type AuthToken struct {
	Id        int64     `json:"id" gorm:"column:id" validate:"required"`
	UserId    int64     `json:"user_id"`
	Token     string    `json:"token" gorm:"column:token" validate:"required"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	ExpiresAt time.Time `json:"expires_at" gorm:"column:expires_at"`
	Revoked   bool      `json:"revoked" gorm:"column:revoked"`
}
