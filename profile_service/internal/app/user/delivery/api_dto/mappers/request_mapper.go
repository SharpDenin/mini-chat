package mappers

import (
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

func ConvertToServiceCreate(u *hDto.CreateUserRequest) *sDto.CreateUserRequest {
	return &sDto.CreateUserRequest{
		Username: u.Name,
		Email:    u.Email,
		Password: u.Password,
	}
}

func ConvertToServiceUpdate(u *hDto.UpdateUserRequest) *sDto.UpdateUserRequest {
	return &sDto.UpdateUserRequest{
		Username: u.Name,
		Email:    u.Email,
		Password: u.Password,
	}
}
