package service_dto

type FriendListResponse struct {
	Friends []FriendView `json:"friends"`
	Total   int          `json:"total"`
}

type FriendView struct {
	Id        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at,omitempty"`
}
