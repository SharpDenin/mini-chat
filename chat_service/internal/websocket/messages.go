package websocket

import "encoding/json"

type MessageType string

const (
	MessagePing     MessageType = "ping"
	MessageChat     MessageType = "chat"
	MessageSystem   MessageType = "system"
	MessagePresence MessageType = "presence"
)

type WSMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}
