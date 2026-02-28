package handler

import (
	"chat_service/internal/pubsub"
	"chat_service/internal/websocket"
	"chat_service/internal/websocket/dto"
	"chat_service/internal/websocket/helper"
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
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
	allowed, err := c.Authz.CanSendDirect(c.Ctx, c.UserId, payload.ToUserId)
	if err != nil {
		logrus.Debug("authz_error")
		return
	}

	//TODO: Send back "Not allowed msg"
	if !allowed {
		logrus.Debug("not_allowed")
		return
	}
	data, _ := json.Marshal(map[string]any{
		"to_user_id":   payload.ToUserId,
		"from_user_id": c.UserId,
		"text":         payload.Text,
	})

	c.Hub.SendToUser(payload.ToUserId, helper.BuildChatWS(data))

	event := pubsub.RedisEvent{
		Type:       "direct",
		InstanceId: c.Hub.InstanceId,
		Data:       data,
	}

	raw, _ := json.Marshal(event)

	_ = c.Hub.Pubsub.Publish(c.Ctx, "chat.direct", raw)
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

	data, _ := json.Marshal(map[string]any{
		"room_id":      payload.RoomId,
		"from_user_id": c.UserId,
		"text":         payload.Text,
	})

	c.Hub.BroadcastToRoom(payload.RoomId, helper.BuildChatWS(data))

	event := pubsub.RedisEvent{
		Type:       "room",
		InstanceId: c.Hub.InstanceId,
		Data:       data,
	}

	raw, _ := json.Marshal(event)

	_ = c.Hub.Pubsub.Publish(c.Ctx, "chat.room", raw)
}
