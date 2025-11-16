package room_models

import "time"

type RoomMember struct {
	Id       int64     `gorm:"primaryKey;autoIncrement;column:id"`
	RoomId   int64     `gorm:"column:room_id;not null;index"`
	UserId   int64     `gorm:"column:user_id;not null;index"`
	JoinedAt time.Time `gorm:"type:timestamp with time zone;default:now()"`
	IsAdmin  bool      `gorm:"column:is_admin;default:false"`
}
