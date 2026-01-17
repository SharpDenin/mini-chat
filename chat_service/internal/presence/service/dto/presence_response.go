package dto

import "time"

type PresenceStatus string

const (
	Online  PresenceStatus = "online"
	Offline PresenceStatus = "offline"
	Idle    PresenceStatus = "idle"
)

type Presence struct {
	UserId   int64
	Status   PresenceStatus
	LastSeen time.Time
}
