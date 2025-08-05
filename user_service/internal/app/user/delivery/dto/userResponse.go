package dto

import (
	"time"
	"user_service/internal/app/user/service/dto"
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

func ConvertToServiceUser(u *dto.GetUserResponse) *UserViewResponse {
	return &UserViewResponse{
		Id:        u.Id,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

func ConvertToServiceList(u *dto.GetUserViewListResponse) *UserViewListResponse {
	resp := &UserViewListResponse{
		Limit:  u.Limit,
		Offset: u.Offset,
		Total:  u.Total,
		UserList:  make([]*UserViewResponse, 0, len(u.UserList)),
	}

	for _, user := range u.UserList {
		resp.UserList = append(resp.UserList, &UserViewResponse{
			Id:        user.Id,
			Name:  	   user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		})
	}

	return resp
}
