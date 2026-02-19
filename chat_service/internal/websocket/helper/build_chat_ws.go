package helper

import (
	"chat_service/internal/websocket/dto"
	"encoding/json"
)

func BuildChatWS(payload []byte) []byte {
	msg, _ := json.Marshal(dto.WSMessage{
		Type:    dto.MessageChat,
		Payload: payload,
	})
	return msg
}
