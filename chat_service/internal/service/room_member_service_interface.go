package service

import (
	"chat_service/internal/models"
	"context"
)

type RoomMemberServiceInterface interface {
	AddMember(ctx context.Context, roomID, userID int64) error
	RemoveMember(ctx context.Context, roomID, userID int64) error
	ListMembers(ctx context.Context, roomID int64) ([]*models.RoomMember, error)
	SetAdmin(ctx context.Context, roomID, userID int64, isAdmin bool) error
	ListUserRooms(ctx context.Context, userID int64) ([]*models.Room, error)
}
