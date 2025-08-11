package mappers

import (
	"profile_service/internal/app/auth/gRPC"
	hDto "profile_service/internal/app/user/delivery/api_dto"
	sDto "profile_service/internal/app/user/service/dto"
)

func ConvertToServiceFilter(f *hDto.UserFilterRequest) *sDto.SearchUserFilter {
	return &sDto.SearchUserFilter{
		Username: f.Name,
		Email:    f.Email,
		Limit:    f.Limit,
		Offset:   f.Offset,
	}
}

func ConvertToRegisterRequest(u *hDto.CreateUserRequest) *gRPC.RegisterRequest {
	return &gRPC.RegisterRequest{
		Username: u.Name,
		Email:    u.Email,
		Password: u.Password,
	}
}

func ConvertToServiceUpdate(u *hDto.UpdateUserRequest) *sDto.UpdateUserRequest {
	return &sDto.UpdateUserRequest{
		Username: u.Name,
		Email:    u.Email,
	}
}

func ConvertToLoginRequest(u *hDto.LoginRequest) *gRPC.LoginRequest {
	return &gRPC.LoginRequest{
		Username: u.Username,
		Password: u.Password,
	}
}
