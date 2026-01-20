package websocket

import (
	"chat_service/internal/presence/service"
	"context"
	"time"

	"github.com/gorilla/websocket"
)

type Connection struct {
	ws       *websocket.Conn
	userId   int64
	connId   int64
	send     chan []byte
	presence service.PresenceService
	hub      *Hub
}

func NewConnection(ws *websocket.Conn, userId int64, presence service.PresenceService, hub *Hub) *Connection {
	return &Connection{
		ws:       ws,
		userId:   userId,
		connId:   time.Now().UnixNano(),
		send:     make(chan []byte, 256),
		presence: presence,
		hub:      hub,
	}
}

func (c *Connection) Start() {
	_ = c.presence.OnConnect(context.Background(), c.userId, c.connId, "web")

	go c.readLoop()
	go c.writeLoop()
}

func (c *Connection) close() {
	c.hub.Unregister(c)
	_ = c.presence.OnDisconnect(context.Background(), c.userId, c.connId)
	_ = c.ws.Close()
}
