package presence_models

import "time"

type PresenceEvent struct {
	UserID   string    `json:"user_id"`
	Status   string    `json:"status"`
	LastSeen time.Time `json:"last_seen"`
}
