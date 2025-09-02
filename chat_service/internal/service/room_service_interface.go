package service

import (
	"chat_service/internal/models"
	"context"
)

type RoomServiceInterface interface {
	CreateRoom(ctx context.Context, name string) (int64, error)
	RenameRoomById(ctx context.Context, roomID int64, name string) error
	DeleteRoomById(ctx context.Context, roomID int64) error
	GetRoomById(ctx context.Context, roomID int64) (*models.Room, error)
	GetRoomList(ctx context.Context, search string, limit, offset int) ([]*models.Room, error)
}
