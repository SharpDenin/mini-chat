package kafka

import "context"

type ProducerInterface interface {
	SendEvent(ctx context.Context, topic, key string, value interface{}) error
}
