package api_dto

type UserFilterRequest struct {
	Name   *string `form:"name" json:"name"`
	Email  *string `form:"email" json:"email"`
	Limit  int     `form:"limit,default=10" json:"limit"`
	Offset int     `form:"offset,default=0" json:"offset"`
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name" binding:"omitempty"`
	Email    *string `json:"email" binding:"omitempty"`
	Password *string `json:"password" binding:"omitempty"`
}
