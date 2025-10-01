package user_mapper

import (
	sDto "profile_service/internal/service/service_dto"
	hDto "profile_service/internal/transport/api_dto"
	"profile_service/pkg/grpc_generated/profile"
)

func ConvertToServiceUser(u *sDto.GetUserResponse) *hDto.UserViewResponse {
	return &hDto.UserViewResponse{
		Id:        u.Id,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

func ConvertToServiceList(u *sDto.GetUserViewListResponse) *hDto.UserViewListResponse {
	resp := &hDto.UserViewListResponse{
		Limit:    u.Limit,
		Offset:   u.Offset,
		Total:    u.Total,
		UserList: make([]*hDto.UserViewResponse, 0, len(u.UserList)),
	}

	for _, user := range u.UserList {
		resp.UserList = append(resp.UserList, &hDto.UserViewResponse{
			Id:        user.Id,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		})
	}
	return resp
}

func ConvertToLoginResponse(g *profile.LoginResponse) *hDto.LoginResponse {
	return &hDto.LoginResponse{
		Token:  g.Token,
		UserId: g.UserId,
	}
}
