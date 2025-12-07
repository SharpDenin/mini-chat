package repo_dto

import "time"

type ConnectionInfoResponse struct {
	ConnId       int64     `json:"conn_id"`
	UserId       int64     `json:"user_id"`
	DeviceType   string    `json:"device_type"`
	ConnectedAt  time.Time `json:"connected_at"`
	LastActivity time.Time `json:"last_activity"`
}
