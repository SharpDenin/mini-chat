package websocket

import "time"

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = pongWait * 9 / 10
	maxMessageSize = 8 * 1024 // 8KB
)

type Hub struct {
	register   chan *Connection
	unregister chan *Connection
	conns      map[int64]map[int64]*Connection
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Connection),
		unregister: make(chan *Connection),
		conns:      make(map[int64]map[int64]*Connection),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			if h.conns[c.userId] == nil {
				h.conns[c.userId] = make(map[int64]*Connection)
			}
			h.conns[c.userId][c.connId] = c

		case c := <-h.unregister:
			if userConns, ok := h.conns[c.userId]; ok {
				delete(userConns, c.connId)
				if len(userConns) == 0 {
					delete(h.conns, c.userId)
				}
			}
		}
	}
}

func (h *Hub) Register(c *Connection) {
	h.register <- c
}

func (h *Hub) Unregister(c *Connection) {
	h.unregister <- c
}
