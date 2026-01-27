package websocket

import (
	"context"
	"encoding/json"
	"time"
)

func (c *Connection) readLoop() {
	defer c.close()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))

	c.ws.SetPongHandler(func(string) error {
		_ = c.presence.OnHeartbeat(context.Background(), c.connId)
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.ws.ReadMessage()
		if err != nil {
			return
		}

		var msg WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue // плохой клиент
		}

		c.router.Route(c.ctx, c, msg)
	}
}
