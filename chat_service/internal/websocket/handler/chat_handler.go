package handler

import (
	"chat_service/internal/websocket"
	"chat_service/internal/websocket/dto"
	"context"
	"encoding/json"
)

func ChatHandler(ctx context.Context, c *websocket.Connection, msg dto.WSMessage) {
	var payload dto.ChatPayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return
	}

	switch payload.Kind {

	case dto.ChatDirect:
		handleDirect(c, payload)

	case dto.ChatRoom:
		handleRoom(c, payload)

	}
}

func handleDirect(c *websocket.Connection, payload dto.ChatPayload) {
	pl, _ := json.Marshal(map[string]any{
		"from_user_id": c.UserId,
		"text":         payload.Text,
	})

	resp, _ := json.Marshal(dto.WSMessage{
		Type:    dto.MessageChat,
		Payload: pl,
	})

	c.Hub.SendToUser(payload.ToUserId, resp)
}

func handleRoom(c *websocket.Connection, payload dto.ChatPayload) {
	if payload.Action == "join" {
		c.Hub.JoinRoom(payload.RoomId, c)
		return
	}

	if payload.Action == "leave" {
		c.Hub.LeaveRoom(payload.RoomId, c)
		return
	}

	pl, _ := json.Marshal(map[string]any{
		"room_id":      payload.RoomId,
		"from_user_id": c.UserId,
		"text":         payload.Text,
	})

	resp, _ := json.Marshal(dto.WSMessage{
		Type:    dto.MessageChat,
		Payload: pl,
	})

	c.Hub.BroadcastToRoom(payload.RoomId, resp)
}
