package models

import (
	"encoding/json"
	"time"
)

// EventType определяет типы событий дружбы
type EventType string

const (
	EventFriendRequestSent      EventType = "friend_request_sent"
	EventFriendRequestAccepted  EventType = "friend_request_accepted"
	EventFriendRequestRejected  EventType = "friend_request_rejected"
	EventFriendRequestCancelled EventType = "friend_request_cancelled"
	EventFriendAdded            EventType = "friend_added"
	EventFriendRemoved          EventType = "friend_removed"
	EventUserBlocked            EventType = "user_blocked"
	EventUserUnblocked          EventType = "user_unblocked"
)

// BaseEvent базовое событие для всех событий
type BaseEvent struct {
	EventId       string    `json:"event_id"`
	EventType     EventType `json:"event_type"`
	UserId        int64     `json:"user_id"`
	Timestamp     time.Time `json:"timestamp"`
	Service       string    `json:"service"`
	Version       string    `json:"version"`
	CorrelationId string    `json:"correlation_id,omitempty"`
}

// FriendRequestSentEvent событие отправки запроса в друзья
type FriendRequestSentEvent struct {
	BaseEvent
	SenderId   int64  `json:"sender_id"`
	ReceiverId int64  `json:"receiver_id"`
	RequestId  int64  `json:"request_id"`
	Message    string `json:"message,omitempty"`
}

// MarshalJSON сериализует событие в JSON
func (e *FriendRequestSentEvent) MarshalJSON() ([]byte, error) {
	type Alias FriendRequestSentEvent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	})
}

// FriendRequestActionEvent событие действия с запросом (accept/reject)
type FriendRequestActionEvent struct {
	BaseEvent
	RequestId int64  `json:"request_id"`
	FriendId  int64  `json:"friend_id"`
	Action    string `json:"action"` // accept/reject
}

// FriendEvent событие изменения списка друзей
type FriendEvent struct {
	BaseEvent
	FriendId int64  `json:"friend_id"`
	Action   string `json:"action"` // add/remove
}

// BlockEvent событие блокировки/разблокировки
type BlockEvent struct {
	BaseEvent
	BlockedId int64  `json:"blocked_id"`
	Action    string `json:"action"` // block/unblock
	Reason    string `json:"reason,omitempty"`
}
