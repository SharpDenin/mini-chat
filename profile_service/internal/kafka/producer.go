package kafka

import "context"

type ProducerInterface interface {
	SendEvent(ctx context.Context, topic, key string, value interface{}) error
}

type KafkaProducer struct {
}

type ProducerMetrics struct {
}

func NewAsyncKafkaProducer(brokers []string, topic string) (ProducerInterface, error) {
	return &KafkaProducer{}, nil
}

func (k KafkaProducer) SendEvent(ctx context.Context, topic, key string, value interface{}) error {
	//TODO implement me
	panic("implement me")
}
