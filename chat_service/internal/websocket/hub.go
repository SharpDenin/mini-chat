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
	register    chan *Connection
	unregister  chan *Connection
	connections map[*Connection]struct{}

	presenceSub service.PresenceSubscriber
}

func NewHub(presenceSub service.PresenceSubscriber) *Hub {
	return &Hub{
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		connections: make(map[*Connection]struct{}),

		presenceSub: presenceSub,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = struct{}{}

		case c := <-h.unregister:
			delete(h.connections, c)

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
