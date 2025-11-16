package room_repository

import (
	"chat_service/internal/room/room_models"
	"context"
)

type RoomRepoInterface interface {
	Create(ctx context.Context, room *room_models.Room) error
	GetRoomById(ctx context.Context, id int64) (*room_models.Room, error)
	GetAll(ctx context.Context, searchFilter string, limit, offset int) ([]*room_models.Room, error)
	Update(ctx context.Context, id int64, room *room_models.Room) error
	Delete(ctx context.Context, id int64) error
}

type RoomMemberRepoInterface interface {
	AddMember(ctx context.Context, roomId, userId int64) error
	GetMemberByUserId(ctx context.Context, roomId, userId int64) (*room_models.RoomMember, error)
	RemoveMember(ctx context.Context, roomId, userId int64) error
	RemoveAllMembers(ctx context.Context, roomId int64) error
	GetMembersByRoom(ctx context.Context, roomId int64) ([]*room_models.RoomMember, error)
	GetRoomsByUserId(ctx context.Context, userId int64) ([]*room_models.Room, error)
	SetAdmin(ctx context.Context, roomId, userId int64, isAdmin bool) error
}
