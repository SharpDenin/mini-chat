package websocket

import (
	"chat_service/internal/presence/service"
	"chat_service/internal/pubsub"
	"chat_service/internal/websocket/dto"
	"chat_service/internal/websocket/helper"
	"context"
	"encoding/json"
	"log"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = pongWait * 9 / 10
	maxMessageSize = 8 * 1024 // 8KB
)

type Hub struct {
	register   chan *Connection
	unregister chan *Connection

	users map[int64]map[*Connection]struct{}

	rooms map[int64]map[*Connection]struct{}

	connections map[*Connection]struct{}

	Pubsub     pubsub.PubSub
	InstanceId string

	presenceSub service.PresenceSubscriber
}

func NewHub(presenceSub service.PresenceSubscriber, pub pubsub.PubSub, instanceId string) *Hub {
	return &Hub{
		register:   make(chan *Connection),
		unregister: make(chan *Connection),

		users: make(map[int64]map[*Connection]struct{}),
		rooms: make(map[int64]map[*Connection]struct{}),

		connections: make(map[*Connection]struct{}),

		Pubsub:     pub,
		InstanceId: instanceId,

		presenceSub: presenceSub,
	}
}

func (h *Hub) Run(ctx context.Context) {

	directCh, err := h.Pubsub.Subscribe(ctx, "chat.direct")
	if err != nil {
		panic(err)
	}

	roomCh, err := h.Pubsub.Subscribe(ctx, "chat.room")
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			return

		case c := <-h.register:
			h.addConnection(c)

		case c := <-h.unregister:
			h.removeConnection(c)

		case evt := <-h.presenceSub:
			h.broadcastPresence(evt)

		case raw := <-directCh:
			h.handleRedisDirect(raw)

		case raw := <-roomCh:
			h.handleRedisRoom(raw)
		}
	}
}

func (h *Hub) RegisterConnection(c *Connection) {
	h.register <- c
}

func (h *Hub) UnregisterConnection(c *Connection) {
	h.unregister <- c
}

func (h *Hub) addConnection(c *Connection) {
	h.connections[c] = struct{}{}

	if h.users[c.UserId] == nil {
		h.users[c.UserId] = make(map[*Connection]struct{})
	}
	h.users[c.UserId][c] = struct{}{}

	log.Printf("[hub] user %d connected, total connections: %d",
		c.UserId, len(h.users[c.UserId]))
}

func (h *Hub) removeConnection(c *Connection) {
	delete(h.connections, c)

	if conns, ok := h.users[c.UserId]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.users, c.UserId)
		}
	}

	log.Printf("[hub] user %d disconnected", c.UserId)
}

func (h *Hub) broadcastPresence(evt service.PresenceEvent) {
	payload, err := json.Marshal(map[string]any{
		"user_id": evt.UserId,
		"event":   evt.Type,
	})
	if err != nil {
		log.Printf("failed to marshal presence payload: %v", err)
		return
	}

	msg, err := json.Marshal(dto.WSMessage{
		Type:    dto.MessagePresence,
		Payload: payload,
	})
	if err != nil {
		log.Printf("failed to marshal presence message: %v", err)
		return
	}

	log.Printf("[presence] broadcasting %s for user %d to %d connections",
		evt.Type, evt.UserId, len(h.connections))

	for c := range h.connections {
		if _, ok := c.Subscribed[evt.UserId]; ok {
			c.Send <- msg
		}
	}
}

func (h *Hub) SendToUser(userId int64, msg []byte) {
	if conns, ok := h.users[userId]; ok {
		for c := range conns {
			c.Send <- msg
		}
	}
}

func (h *Hub) JoinRoom(roomId int64, c *Connection) {
	if h.rooms[roomId] == nil {
		h.rooms[roomId] = make(map[*Connection]struct{})
	}
	h.rooms[roomId][c] = struct{}{}
}

func (h *Hub) LeaveRoom(roomId int64, c *Connection) {
	if members, ok := h.rooms[roomId]; ok {
		delete(members, c)
		if len(members) == 0 {
			delete(h.rooms, roomId)
		}
	}
}

func (h *Hub) BroadcastToRoom(roomId int64, msg []byte) {
	if members, ok := h.rooms[roomId]; ok {
		for c := range members {
			c.Send <- msg
		}
	}
}

func (h *Hub) handleRedisDirect(raw []byte) {
	var evt pubsub.RedisEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		return
	}

	if evt.InstanceId != h.InstanceId {
		return
	}

	var payload struct {
		ToUserId   int64  `json:"to_user_id"`
		FromUserId int64  `json:"from_user_id"`
		Text       string `json:"text"`
	}

	if err := json.Unmarshal(evt.Data, &payload); err != nil {
		return
	}

	h.SendToUser(payload.ToUserId, helper.BuildChatWS(evt.Data))
}

func (h *Hub) handleRedisRoom(raw []byte) {
	var evt pubsub.RedisEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		return
	}

	if evt.InstanceId == h.InstanceId {
		return
	}

	var payload struct {
		RoomId     int64  `json:"room_id"`
		FromUserId int64  `json:"from_user_id"`
		Text       string `json:"text"`
	}

	if err := json.Unmarshal(evt.Data, &payload); err != nil {
		return
	}

	h.BroadcastToRoom(payload.RoomId, helper.BuildChatWS(evt.Data))
}
