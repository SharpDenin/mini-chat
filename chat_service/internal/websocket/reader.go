package websocket

import (
	"context"
	"time"
)

func (c *Connection) readLoop() {
	defer c.close()

	c.ws.SetReadLimit(512)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		_ = c.presence.OnHeartbeat(context.Background(), c.connId)
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			return
		}

		// TODO: decode message & route
		_ = msg
	}
}
