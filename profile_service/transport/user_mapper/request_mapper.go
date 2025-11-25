package user_mapper

import (
	"profile_service/internal/user/service/service_dto"
	"profile_service/pkg/grpc_generated/profile"
	hDto "profile_service/transport/api_dto"
)

func ConvertToServiceFilter(f *hDto.UserFilterRequest) *service_dto.SearchUserFilter {
	return &service_dto.SearchUserFilter{
		Username: f.Name,
		Email:    f.Email,
		Limit:    f.Limit,
		Offset:   f.Offset,
	}
}

func ConvertToRegisterRequest(u *hDto.CreateUserRequest) *profile.RegisterRequest {
	return &profile.RegisterRequest{
		Username: u.Name,
		Email:    u.Email,
		Password: u.Password,
	}
}

func ConvertToServiceUpdate(u *hDto.UpdateUserRequest) *service_dto.UpdateUserRequest {
	return &service_dto.UpdateUserRequest{
		Username: u.Name,
		Email:    u.Email,
	}
}

func ConvertToLoginRequest(u *hDto.LoginRequest) *profile.LoginRequest {
	return &profile.LoginRequest{
		Username: u.Username,
		Password: u.Password,
	}
}
