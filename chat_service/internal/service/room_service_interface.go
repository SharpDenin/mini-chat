package service

import (
	"chat_service/internal/models"
	"context"
)

type RoomServiceInterface interface {
	CreateRoom(ctx context.Context, name string) (int64, error)
	RenameRoom(ctx context.Context, roomID int64, newName string) error
	DeleteRoom(ctx context.Context, roomID int64) error
	GetRoom(ctx context.Context, roomID int64) (*models.Room, error)
	ListRooms(ctx context.Context, search string, limit, offset int) ([]*models.Room, error)
}
