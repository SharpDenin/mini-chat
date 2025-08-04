package dto

import (
	"time"
	"user_service/internal/app/user/entities/model"
)

func ToUserCreateDto(user *model.User) *UserCreateDTO {
	return &UserCreateDTO{
		Name:  user.Username,
		Email: user.Email,
	}
}

func ToUserViewDTO(user *model.User) *UserViewDTO {
	return &UserViewDTO{
		Name:      user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

// UserViewDTO - view
type UserViewDTO struct {
	Name      string    `json:"name" binding:"required, min=1, max=50"`
	Email     string    `json:"email" binding:"required,email"`
	CreatedAt time.Time `json:"created_at" binding:"required"`
}

// UserCreateDTO - create
type UserCreateDTO struct {
	Name  string `json:"name" binding:"required, min=1, max=50"`
	Email string `json:"email" binding:"required,email"`
}

type UserListDTO struct {
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	List   []*UserViewDTO `json:"list"`
}

type SearchUserFilterDTO struct {
	Username string
	Email    string
	SortBy   string
	Limit    int
	Offset   int
}
