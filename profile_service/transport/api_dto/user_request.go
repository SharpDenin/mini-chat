package api_dto

type UserFilterRequest struct {
	Name   string `form:"name" json:"name" validate:"omitempty,min=2,max=50"`
	Email  string `form:"email" json:"email" validate:"omitempty,email"`
	Status string `form:"status" json:"status" validate:"omitempty"`
	Limit  int    `form:"limit,default=10" json:"limit"`
	Offset int    `form:"offset,default=0" json:"offset"`
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name" binding:"omitempty,min=2,max=50"`
	Email *string `json:"email" binding:"omitempty,email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
