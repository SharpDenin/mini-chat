package dto

type BulkPresenceResponse struct {
	Presences map[int64]PresenceResponse `json:"presences"`
	Errors    map[int64]PresenceError    `json:"errors,omitempty"`
}
