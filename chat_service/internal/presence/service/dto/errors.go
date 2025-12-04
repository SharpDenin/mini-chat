package dto

import "errors"

var (
	ErrInvalidUserId      = errors.New("invalid user id")
	ErrUserNotFound       = errors.New("user not found")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrServiceUnavailable = errors.New("presence service unavailable")
	ErrAlreadyOnline      = errors.New("user is already online")
	ErrAlreadyOffline     = errors.New("user is already offline")
)

type PresenceError struct {
	Err     error
	UserId  int64
	Action  string
	Context map[string]interface{}
}

func (e *PresenceError) Error() string {
	return e.Err.Error()
}

func (e *PresenceError) Unwrap() error {
	return e.Err
}
