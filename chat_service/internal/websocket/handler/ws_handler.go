package handler

import (
	"chat_service/internal/authz"
	"chat_service/internal/helpers"
	"chat_service/internal/presence/service"
	webS "chat_service/internal/websocket"
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

func NewWSHandler(ctx context.Context, router *webS.Router, hub *webS.Hub, presence service.PresenceService, authz authz.AuthServiceInterface) http.HandlerFunc {
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

		conn := webS.NewConnection(ws, userId, presence, ctx, router, hub, authz)
		hub.RegisterConnection(conn)
		conn.Start()
	}
}
