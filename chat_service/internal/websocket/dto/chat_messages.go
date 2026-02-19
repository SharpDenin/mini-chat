package dto

type ChatKind string

const (
	ChatDirect ChatKind = "direct"
	ChatRoom   ChatKind = "room"
)

type ChatPayload struct {
	Kind     ChatKind `json:"kind"`
	ToUserId int64    `json:"to_user_id,omitempty"`
	RoomId   int64    `json:"room_id,omitempty"`
	Text     string   `json:"text"`

	Action string `json:"action,omitempty"`
}
