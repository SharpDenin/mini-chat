package websocket

import (
	"context"
	"log"
)

type HandlerFunc func(ctx context.Context, c *Connection, msg WSMessage)

type Router struct {
	handlers map[MessageType]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		handlers: make(map[MessageType]HandlerFunc),
	}
}

func (r *Router) Register(t MessageType, h HandlerFunc) {
	r.handlers[t] = h
}

func (r *Router) Route(ctx context.Context, c *Connection, msg WSMessage) {
	h, ok := r.handlers[msg.Type]
	if !ok {
		log.Printf("[ws] unknown message type: %s", msg.Type)
		return
	}

	h(ctx, c, msg)
}
