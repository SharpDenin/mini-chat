package service

import (
	"chat_service/internal/service/dto"
	"context"
)

type RoomServiceInterface interface {
	CreateRoom(ctx context.Context, name string) (int64, error)
	RenameRoomById(ctx context.Context, roomID int64, name string) error
	DeleteRoomById(ctx context.Context, roomID int64) error
	GetRoomById(ctx context.Context, roomID int64) (*dto.GetRoomResponse, error)
	GetRoomList(ctx context.Context, filter *dto.SearchFilter) ([]*dto.GetRoomResponse, error)
}
