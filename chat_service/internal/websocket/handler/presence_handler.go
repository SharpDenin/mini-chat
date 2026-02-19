package handler

import (
	"chat_service/internal/websocket"
	"chat_service/internal/websocket/dto"
	"context"
	"encoding/json"
)

func PresenceHandler(ctx context.Context, c *websocket.Connection, msg dto.WSMessage) {
	var payload dto.PresencePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return
	}

	switch payload.Cmd {

	case dto.CmdSubscribe:
		handleSubscribe(c, payload.UserIds)

	case dto.CmdUnsubscribe:
		handleUnsubscribe(c, payload.UserIds)

	case dto.CmdGetOnlineFriends:
		handleGetOnlineFriends(ctx, c, payload.UserIds)

	default:

	}
}

func handleSubscribe(c *websocket.Connection, userIDs []int64) {
	for _, id := range userIDs {
		c.Subscribed[id] = struct{}{}
	}
}

func handleUnsubscribe(c *websocket.Connection, userIDs []int64) {
	for _, id := range userIDs {
		delete(c.Subscribed, id)
	}
}

func handleGetOnlineFriends(ctx context.Context, c *websocket.Connection, userIds []int64) {
	online := c.Presence.GetOnlineFriends(ctx, c.UserId, userIds)

	payload, _ := json.Marshal(map[string]any{
		"cmd":   "online_list",
		"users": online,
	})

	resp, _ := json.Marshal(dto.WSMessage{
		Type:    dto.MessagePresence,
		Payload: payload,
	})

	c.Send <- resp
}
