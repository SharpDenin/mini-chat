package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

// FriendshipHistory - история всех событий дружбы (для Kafka и аналитики)
type FriendshipHistory struct {
	Id        int64           `gorm:"primaryKey;autoIncrement;column:id"`
	EventType string          `gorm:"column:event_type;type:varchar(50);not null;index"` // request_sent, request_accepted, request_rejected, request_cancelled, unfriended, blocked, unblocked
	UserId    int64           `gorm:"column:user_id;not null;index:idx_history_user"`
	TargetId  int64           `gorm:"column:target_id;not null;index:idx_history_target"`
	OldStatus *string         `gorm:"column:old_status;type:varchar(20)"`
	NewStatus *string         `gorm:"column:new_status;type:varchar(20)"`
	Metadata  json.RawMessage `gorm:"column:metadata;type:jsonb"` // дополнительная информация (сообщение, причина блокировки и т.д.)
	RequestId *int64          `gorm:"column:request_id;index"`    // ссылка на FriendRequest если применимо
	CreatedAt time.Time       `gorm:"column:created_at;type:timestamp with time zone;default:now();index"`
}

func (fh *FriendshipHistory) BeforeCreate(tx *gorm.DB) error {
	fh.CreatedAt = time.Now().UTC()
	return nil
}

// TableName задает имя таблицы для истории
func (FriendshipHistory) TableName() string {
	return "friendship_history"
}
