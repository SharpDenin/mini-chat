package repo_dto

import "time"

// UserPresence - DTO для возврата из репозитория
type UserPresence struct {
	UserId   string
	Online   bool
	Status   string
	LastSeen time.Time
}
