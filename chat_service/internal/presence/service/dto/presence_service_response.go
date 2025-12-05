package dto

import (
	"time"
)

type UserStatus string

const (
	StatusOnline  UserStatus = "online"
	StatusOffline UserStatus = "offline"
	StatusIdle    UserStatus = "idle"
)

type PresenceResponse struct {
	UserId      int64      `json:"user_id"`
	Status      UserStatus `json:"status"`
	LastSeen    time.Time  `json:"last_seen"`
	Connections int        `json:"connections,omitempty"`
	DeviceType  string     `json:"device_type,omitempty"`
}

type BulkPresenceResponse struct {
	Presences map[int64]PresenceResponse `json:"presences"`
	Errors    map[int64]PresenceError    `json:"errors,omitempty"`
}

type StatusChangeEvent struct {
	UserId    int64      `json:"user_id"`
	OldStatus UserStatus `json:"old_status"`
	NewStatus UserStatus `json:"new_status"`
	Timestamp time.Time  `json:"timestamp"`
	Source    string     `json:"source"`
}
