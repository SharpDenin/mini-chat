package dto

import "time"

type StatusChangeResponse struct {
	UserId    int64      `json:"user_id"`
	OldStatus UserStatus `json:"old_status"`
	NewStatus UserStatus `json:"new_status"`
	Timestamp time.Time  `json:"timestamp"`
	Source    string     `json:"source"`
}
