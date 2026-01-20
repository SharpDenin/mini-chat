package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

func (c *Connection) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			_ = c.ws.WriteMessage(websocket.TextMessage, msg)

		case <-ticker.C:
			_ = c.ws.WriteMessage(websocket.PingMessage, nil)
		}
	}
}
