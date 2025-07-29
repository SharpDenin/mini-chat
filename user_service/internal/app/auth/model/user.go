package model

import "time"

type User struct {
	Id        string    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username" validate:"required, min=1, max=50"`
	Email     string    `json:"email" db:"email" validate:"required,email"`
	Password  string    `json:"password" db:"password" validate:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
