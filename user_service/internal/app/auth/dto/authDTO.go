package dto

type RegisterRequestDTO struct {
	Name     string `json:"name" binding:"required, min=1, max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required, min=8"`
}

type LoginRequestDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required, min=8"`
}

type AuthResponseDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
