package service

type PresenceEventType string

const (
	EventUserOnline  PresenceEventType = "user_online"
	EventUserOffline PresenceEventType = "user_offline"
)

type PresenceEvent struct {
	Type   PresenceEventType
	UserId int64
}
