package service_dto

import (
	"time"
)

type GetUserResponse struct {
	Id        int64
	Name      string
	Email     string
	Password  string
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
	SortBy   string
	Limit    int
	Offset   int
}
