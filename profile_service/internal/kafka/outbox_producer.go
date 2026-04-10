package kafka

import (
	"context"
	"encoding/json"
	"gorm.io/gorm"
	"profile_service/internal/kafka/models"
	"time"
)

type OutboxProducer struct {
	db        *gorm.DB
	producer  ProducerInterface
	ticker    *time.Ticker
	stopChan  chan struct{}
	batchSize int
}

func NewOutboxProducer(db *gorm.DB, producer ProducerInterface, batchSize int) *OutboxProducer {
	p := &OutboxProducer{
		db:        db,
		producer:  producer,
		ticker:    time.NewTicker(5 * time.Second),
		stopChan:  make(chan struct{}),
		batchSize: batchSize,
	}

	go p.processOutbox()

	return p
}

func (p *OutboxProducer) SendEvent(ctx context.Context, topic, key string, value interface{}) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	headers, err := json.Marshal(map[string]string{
		"topic": topic,
		"key":   key,
	})
	if err != nil {
		return err
	}

	var eventType string
	if event, ok := value.(interface{ GetEventType() string }); ok {
		eventType = event.GetEventType()
	} else {
		eventType = "unknown"
	}

	msg := &models.OutboxMessage{
		EventType:   eventType,
		AggregateId: key,
		Payload:     payload,
		Headers:     headers,
		Status:      "pending",
	}

	return p.db.WithContext(ctx).Create(msg).Error
}

func (p *OutboxProducer) processOutbox() {
	for {
		select {
		case <-p.ticker.C:
			p.sendPendingMessages()
		case <-p.stopChan:
			p.ticker.Stop()
			return
		}
	}
}

func (p *OutboxProducer) sendPendingMessages() {
	var messages []models.OutboxMessage

	err := p.db.
		Where("status = ?", "pending").
		Order("created_at ASC").
		Limit(p.batchSize).
		Find(&messages).Error

	if err != nil {
		return
	}

	for _, msg := range messages {
		var headers map[string]string
		json.Unmarshal(msg.Headers, &headers)

		err := p.producer.SendEvent(context.Background(),
			headers["topic"],
			headers["key"],
			msg.Payload)

		if err != nil {
			p.db.Model(&msg).Updates(map[string]interface{}{
				"status":      "failed",
				"retry_count": msg.RetryCount + 1,
				"updated_at":  time.Now(),
			})
			continue
		}

		now := time.Now()
		p.db.Model(&msg).Updates(map[string]interface{}{
			"status":     "sent",
			"sent_at":    &now,
			"updated_at": now,
		})
	}
}

func (p *OutboxProducer) Close() {
	close(p.stopChan)
}
