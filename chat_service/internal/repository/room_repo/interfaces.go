package room_repo

import (
	"chat_service/internal/models"
	"context"
)

type RoomRepoInterface interface {
	Create(ctx context.Context, room *models.Room) error
	GetById(ctx context.Context, id int64) (*models.Room, error)
	GetAll(ctx context.Context, searchFilter string) ([]*models.Room, error)
	Update(ctx context.Context, id int64, room *models.Room) error
	Delete(ctx context.Context, id int64) error
}

type RoomMemberRepoInterface interface {
	AddMember(ctx context.Context, roomId, userId int64) error
	RemoveMember(ctx context.Context, roomId, userId int64) error
	GetMembersByRoom(ctx context.Context, roomId int64) ([]*models.RoomMember, error)
	GetRoomsByUserId(ctx context.Context, userId int64) ([]*models.Room, error)
	SetAdmin(ctx context.Context, roomId, userId int64, isAdmin bool) error
}
