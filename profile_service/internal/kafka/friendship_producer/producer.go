package friendship_producer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

type ProducerInterface interface {
	SendEvent(ctx context.Context, topic, key string, value interface{}) error
	Close() error
}

type AsyncKafkaProducer struct {
	producer    sarama.AsyncProducer
	topic       string
	successChan chan *sarama.ProducerMessage
	errorChan   chan *sarama.ProducerError
	wg          sync.WaitGroup
	stopChan    chan struct{}
	metrics     *ProducerMetrics
}

type ProducerMetrics struct {
	MessagesSent   int64
	MessagesFailed int64
	lastError      time.Time
	mu             sync.RWMutex
}

func NewKafkaProducer(brokers []string, topic string) (ProducerInterface, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = 200 * time.Millisecond
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Flush.Messages = 100
	config.Producer.Flush.MaxMessages = 500

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create async friendship_producer: %w", err)
	}

	k := &AsyncKafkaProducer{
		producer:    producer,
		topic:       topic,
		successChan: make(chan *sarama.ProducerMessage, 100),
		errorChan:   make(chan *sarama.ProducerError, 100),
		stopChan:    make(chan struct{}),
		metrics:     &ProducerMetrics{},
	}

	go k.handleSuccesses()
	go k.handleErrors()

	return k, nil
}

func (k *AsyncKafkaProducer) SendEvent(ctx context.Context, topic, key string, value interface{}) error {
	if topic == "" {
		topic = k.topic
	}

	data, err := json.Marshal(value)
	if err != nil {
		k.metrics.mu.Lock()
		k.metrics.MessagesFailed++
		k.metrics.mu.Unlock()

		return fmt.Errorf("failed to marshal event: %w", err)
	}

	headers := []sarama.RecordHeader{
		{
			Key:   []byte("content-type"),
			Value: []byte("application/json"),
		},
		{
			Key:   []byte("timestamp"),
			Value: []byte(time.Now().UTC().Format(time.RFC3339)),
		},
		{
			Key:   []byte("message-id"),
			Value: []byte(uuid.New().String()),
		},
	}

	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.StringEncoder(key),
		Value:   sarama.ByteEncoder(data),
		Headers: headers,
	}

	select {
	case k.producer.Input() <- msg:
		k.metrics.mu.Lock()
		k.metrics.MessagesSent++
		k.metrics.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (k *AsyncKafkaProducer) GetMetrics() ProducerMetrics {
	k.metrics.mu.RLock()
	defer k.metrics.mu.RUnlock()
	return *k.metrics
}

func (k *AsyncKafkaProducer) Close() error {
	close(k.stopChan)

	k.wg.Wait()

	return k.producer.Close()
}

func (k *AsyncKafkaProducer) handleSuccesses() {
	for msg := range k.producer.Successes() {
		log.Printf("Message sent successfully - topic: %s, partition: %d, offset: %d",
			msg.Topic, msg.Partition, msg.Offset)
	}
}

func (k *AsyncKafkaProducer) handleErrors() {
	for err := range k.producer.Errors() {
		k.metrics.mu.Lock()
		k.metrics.MessagesFailed++
		k.metrics.lastError = time.Now()
		k.metrics.mu.Unlock()

		log.Printf("Failed to send message: %v, topic: %s, key: %s",
			err.Err, err.Msg.Topic, err.Msg.Key)

		go k.retryFailedMessage(err.Msg)
	}
}

func (k *AsyncKafkaProducer) retryFailedMessage(msg *sarama.ProducerMessage) {
	backoff := time.Second
	maxBackoff := time.Minute

	for i := 0; i < 3; i++ {
		time.Sleep(backoff)

		select {
		case k.producer.Input() <- msg:
			log.Printf("Message retry %d successful", i+1)
			return
		default:
			backoff = min(backoff*2, maxBackoff)
		}
	}

	log.Printf("Failed to send message after 3 retries, topic: %s, key: %s", msg.Topic, msg.Key)
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
