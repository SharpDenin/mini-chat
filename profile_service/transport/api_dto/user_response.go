package api_dto

import (
	"time"
)

type UserStatus string

const (
	StatusOnline  UserStatus = "online"
	StatusOffline UserStatus = "offline"
	StatusIdle    UserStatus = "idle"
)

type UserCreateResponse struct {
	Id        int64 `json:"id"`
	CreatedAt int64 `json:"created_at"`
}

type UserViewResponse struct {
	Id        int64      `json:"id"`
	Name      string     `json:"username"`
	Email     string     `json:"email"`
	Status    UserStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

type UserViewListResponse struct {
	Total    int                 `json:"total"`
	Limit    int                 `json:"limit"`
	Offset   int                 `json:"offset"`
	UserList []*UserViewResponse `json:"users"`
}

type LoginResponse struct {
	Token  string     `json:"token"`
	UserId string     `json:"user_id"`
	Status UserStatus `json:"status"`
}
