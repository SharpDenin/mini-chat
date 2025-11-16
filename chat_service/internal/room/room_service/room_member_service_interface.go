package room_service

import (
	"chat_service/internal/room/room_service/dto"
	"context"
)

type RoomMemberServiceInterface interface {
	AddMember(ctx context.Context, roomId, userId int64) error
	RemoveMember(ctx context.Context, roomId, userId int64) error
	ListMembers(ctx context.Context, roomId int64) ([]*dto.GetRoomMemberResponse, error)
	SetAdmin(ctx context.Context, roomId, userId int64, isAdmin bool) error
	ListUserRooms(ctx context.Context, userId int64) ([]*dto.GetRoomResponse, error)
}
