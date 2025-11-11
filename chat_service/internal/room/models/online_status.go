package models

type OnlineStatus struct {
	Id       int64  `gorm:"primary_key;autoIncrement;column:id"`
	UserId   int64  `gorm:"column:user_id;not null;index"`
	Status   string `gorm:"column:status;type:varchar(50);not null"`
	LastSeen int64  `gorm:"column:last_seen;not_null"`
}
