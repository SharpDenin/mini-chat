package dto

import "time"

type GetUserResponse struct {
	Name      string
	Email     string
	CreatedAt time.Time
}

type GetUserViewListResponse struct {
	UserList []*GetUserResponse `json:"users"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
	Total    int                `json:"total"`
}

type SearchUserFilter struct {
	Username string
	Email    string
	SortBy   string
	Limit    int
	Offset   int
}
