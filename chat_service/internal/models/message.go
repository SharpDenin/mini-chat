package models

type Message struct {
	Id     int64  `gorm:"primaryKey;autoIncrement;column:id"`
	UserId int64  `gorm:"column:user_id;not null;index"`
	RoomId int64  `gorm:"column:room_id;not null;index"`
	Text   string `gorm:"type:text;not null;column:text"`
	SentAt int64  `gorm:"column:sent_at;not null"`
}
