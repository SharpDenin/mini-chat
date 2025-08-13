package models

type Room struct {
	Id          int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name        string `gorm:"type:varchar(255);not_null;column:name"`
	UserCount   int64  `gorm:"column:user_count;not null;default:0"`
	OnlineUsers string `gorm:"type:text;column:online_users"`
	LastMessage string `gorm:"type:text;column:last_message"`
}
