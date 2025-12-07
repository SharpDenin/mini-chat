package dto

import "time"

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
