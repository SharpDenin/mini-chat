package websocket

import (
	"chat_service/internal/websocket/dto"
	"context"
	"log"
)

type HandlerFunc func(ctx context.Context, c *Connection, msg dto.WSMessage)

type Router struct {
	handlers map[dto.MessageType]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		handlers: make(map[dto.MessageType]HandlerFunc),
	}
}

func (r *Router) Register(t dto.MessageType, h HandlerFunc) {
	r.handlers[t] = h
}

func (r *Router) Route(ctx context.Context, c *Connection, msg dto.WSMessage) {
	log.Printf("[ws] incoming type: %s", msg.Type)

	h, ok := r.handlers[msg.Type]
	if !ok {
		log.Printf("[ws] unknown message type: %s", msg.Type)
		return
	}

	h(ctx, c, msg)
}
