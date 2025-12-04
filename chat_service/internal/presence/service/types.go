package service

import (
	sDto "chat_service/internal/presence/service/dto"
	"context"
	"time"
)

type Config struct {
	MaxConnectionsPerUser int
	RateLimitPerUser      time.Duration
	CleanupInterval       time.Duration
	IdleThreshold         time.Duration
}

type MetricsCollector interface {
	IncStatusChange(userId int64, from, to string)
	IncConnectionOpened(userId int64, op string)
	ObserveLatency(method string, duration time.Duration)
}

type StatusNotifier interface {
	NotifyStatusChange(ctx context.Context, event sDto.StatusChangeEvent) error
}
