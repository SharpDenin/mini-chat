package service_dto

import (
	"time"
)

type UserStatus string

const (
	StatusOnline  UserStatus = "online"
	StatusOffline UserStatus = "offline"
	StatusUnknown UserStatus = "unknown"
)

type GetUserResponse struct {
	Id        int64
	Name      string
	Email     string
	Password  string
	Status    UserStatus
	CreatedAt time.Time
}

type GetUserViewListResponse struct {
	UserList []*GetUserResponse
	Limit    int
	Offset   int
	Total    int
}

type SearchUserFilter struct {
	Username string
	Email    string
	Status   UserStatus
	SortBy   string
	Limit    int
	Offset   int
}
