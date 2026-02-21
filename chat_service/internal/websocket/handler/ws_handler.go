package handler

import (
	"chat_service/internal/authz"
	"chat_service/internal/presence/service"
	webS "chat_service/internal/websocket"
	"chat_service/middleware_chat"
	"chat_service/pkg/grpc_generated/profile"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Реализовать полноценно
	},
}

func NewWSHandler(ctx context.Context, router *webS.Router, hub *webS.Hub,
	presence service.PresenceService, authz authz.AuthServiceInterface,
	profileClient middleware_chat.ProfileClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		grpcCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		resp, err := profileClient.ValidateToken(grpcCtx, &profile.TokenRequest{
			Token: token,
		})

		if err != nil || !resp.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		userId, _ := strconv.ParseInt(resp.UserId, 10, 64)

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		conn := webS.NewConnection(ws, userId, presence, ctx, router, hub, authz)
		hub.RegisterConnection(conn)
		conn.Start()
	}
}
