package model

import "time"

type AuthToken struct {
	Id        string    `json:"id" db:"id" validate:"required"`
	UserId    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token" validate:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Revoked   bool      `json:"revoked" db:"revoked"`
}
