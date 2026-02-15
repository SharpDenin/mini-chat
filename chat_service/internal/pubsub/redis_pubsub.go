package pubsub

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisPubSub struct {
	client *redis.Client
}

func NewRedisPubSub(client *redis.Client) *RedisPubSub {
	return &RedisPubSub{client: client}
}

func (r *RedisPubSub) Publish(ctx context.Context, channel string, payload []byte) error {
	return r.client.Publish(ctx, channel, payload).Err()
}

func (r *RedisPubSub) Subscribe(ctx context.Context, channel string) (<-chan []byte, error) {
	pubsub := r.client.Subscribe(ctx, channel)

	ch := make(chan []byte, 32)

	go func() {
		defer close(ch)

		for msg := range pubsub.Channel() {
			ch <- []byte(msg.Payload)
		}
	}()

	return ch, nil
}
