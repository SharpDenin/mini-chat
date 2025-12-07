package repo_dto

import "time"

type UserPresenceResponse struct {
	UserId   int64
	Online   bool
	Status   string
	LastSeen time.Time
}
