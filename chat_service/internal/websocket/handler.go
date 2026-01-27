package websocket

import (
	"chat_service/internal/helpers"
	"chat_service/internal/presence/service"
	"context"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Реализовать полноценно
	},
}

func NewWSHandler(ctx context.Context, router *Router, hub *Hub, presence service.PresenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := helpers.GetUserIdFromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		conn := NewConnection(ws, userId, presence, ctx, router, hub)
		hub.Register(conn)
		conn.Start()
	}
}
