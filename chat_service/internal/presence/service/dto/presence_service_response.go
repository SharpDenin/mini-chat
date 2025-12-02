package dto

import "time"

type UserStatus string

const (
	StatusOnline  UserStatus = "online"
	StatusOffline UserStatus = "offline"
)

type PresenceResponse struct {
	UserId      string     `json:"user_id"`
	Status      UserStatus `json:"status"`
	LastSeen    time.Time  `json:"last_seen"`
	Connections int        `json:"connections,omitempty"`
	DeviceType  string     `json:"device_type,omitempty"`
}

type BulkPresenceResponse struct {
	Presences map[string]PresenceResponse `json:"presences"`
	Errors    map[string]string           `json:"errors,omitempty"`
}

type StatusChangeEvent struct {
	UserId    string     `json:"user_id"`
	OldStatus UserStatus `json:"old_status"`
	NewStatus UserStatus `json:"new_status"`
	Timestamp time.Time  `json:"timestamp"`
	Source    string     `json:"source"`
}
