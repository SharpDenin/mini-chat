package models

import (
	"encoding/json"
	"time"
)

type OutboxMessage struct {
	Id          int64           `gorm:"primaryKey;autoIncrement"`
	EventType   string          `gorm:"column:event_type;type:varchar(100);index"`
	AggregateId string          `gorm:"column:aggregate_id;type:varchar(100);index"`
	Payload     json.RawMessage `gorm:"column:payload;type:jsonb"`
	Headers     json.RawMessage `gorm:"column:headers;type:jsonb"`
	Status      string          `gorm:"column:status;type:varchar(20);default:'pending';index"` // pending, sent, failed
	RetryCount  int             `gorm:"column:retry_count;default:0"`
	CreatedAt   time.Time       `gorm:"column:created_at;default:now()"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;default:now()"`
	SentAt      *time.Time      `gorm:"column:sent_at"`
}

func (OutboxMessage) TableName() string {
	return "outbox_messages"
}
