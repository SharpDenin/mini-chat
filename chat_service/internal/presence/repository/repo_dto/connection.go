package repo_dto

import "time"

type Connection struct {
	ConnId       int64
	UserId       int64
	Device       string
	ConnectedAt  time.Time
	LastActivity time.Time
}
