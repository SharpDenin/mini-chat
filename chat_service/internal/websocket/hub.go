package websocket

import (
	"chat_service/internal/presence/service"
	"chat_service/internal/websocket/dto"
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

	presenceSub service.PresenceSubscriber
}

func NewHub(presenceSub service.PresenceSubscriber) *Hub {
	return &Hub{
		register:   make(chan *Connection),
		unregister: make(chan *Connection),

		users: make(map[int64]map[*Connection]struct{}),
		rooms: make(map[int64]map[*Connection]struct{}),

		connections: make(map[*Connection]struct{}),

		presenceSub: presenceSub,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = struct{}{}

			if h.users[c.UserId] == nil {
				h.users[c.UserId] = make(map[*Connection]struct{})
			}
			h.users[c.UserId][c] = struct{}{}

		case c := <-h.unregister:
			delete(h.connections, c)

			if conns, ok := h.users[c.UserId]; ok {
				delete(conns, c)
				if len(conns) == 0 {
					delete(h.users, c.UserId)
				}
			}

			for roomId, members := range h.rooms {
				delete(members, c)
				if len(members) == 0 {
					delete(h.rooms, roomId)
				}
			}

		case evt := <-h.presenceSub:
			h.broadcastPresence(evt)
		}
	}
}

func (h *Hub) Register(c *Connection) {
	h.register <- c
}

func (h *Hub) Unregister(c *Connection) {
	h.unregister <- c
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
