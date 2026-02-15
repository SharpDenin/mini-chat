package websocket

import (
	"chat_service/internal/presence/service"
	"context"
	"time"

	"github.com/gorilla/websocket"
)

type Connection struct {
	ws   *websocket.Conn
	Send chan []byte

	UserId int64
	connId int64

	Ctx      context.Context
	router   *Router
	Presence service.PresenceService
	Hub      *Hub

	Subscribed map[int64]struct{}
}

func NewConnection(ws *websocket.Conn, userId int64, presence service.PresenceService,
	ctx context.Context, router *Router, hub *Hub) *Connection {
	return &Connection{
		ws:   ws,
		Send: make(chan []byte, 256),

		UserId: userId,
		connId: time.Now().UnixNano(),

		Presence: presence,
		Ctx:      ctx,
		router:   router,
		Hub:      hub,

		Subscribed: make(map[int64]struct{}),
	}
}

func (c *Connection) Start() {
	_ = c.Presence.OnConnect(context.Background(), c.UserId, c.connId, "web")

	go c.readLoop()
	go c.writeLoop()
}

func (c *Connection) close() {
	c.Hub.UnregisterConnection(c)
	_ = c.Presence.OnDisconnect(context.Background(), c.UserId, c.connId)
	_ = c.ws.Close()
}
