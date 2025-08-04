package dto

type UserFilterRequest struct {
	Name   *string `form:"name"`
	Email  *string `form:"email"`
	Limit  int     `form:"limit,default=10"`
	Offset int     `form:"offset,default=0"`
}

type UserCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserUpdateRequest struct {
	Name     *string `json:"name" binding:"omitempty"`
	Email    *string `json:"email" binding:"omitempty"`
	Password *string `json:"password" binding:"omitempty"`
}
