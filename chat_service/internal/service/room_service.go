package service

import "chat_service/pkg/grpc_client"

type RoomService struct {
	profileClient *grpc_client.ProfileClient
}

func NewRoomService(profileClient *grpc_client.ProfileClient) RoomServiceInterface {
	return &RoomService{
		profileClient: profileClient,
	}
}
