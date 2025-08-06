package dto

import (
	"time"
)

type UserCreateResponse struct {
	Id        int64 `json:"id"`
	CreatedAt int64 `json:"created_at"`
}

type UserViewResponse struct {
	Id        int64     `json:"id"`
	Name      string    `json:"username" `
	Email     string    `json:"email" `
	CreatedAt time.Time `json:"created_at"`
}

type UserViewListResponse struct {
	UserList []*UserViewResponse `json:"users"`
	Limit    int                 `json:"limit"`
	Offset   int                 `json:"offset"`
	Total    int                 `json:"total"`
}
