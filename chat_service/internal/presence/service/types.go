package service

import (
	sDto "chat_service/internal/presence/service/dto"
	"context"
	"time"
)

type MetricsCollector interface {
	IncStatusChange(userId int64, from, to string)
	IncConnectionOpened(userId int64, op string)
	ObserveLatency(method string, duration time.Duration)
}

type StatusNotifier interface {
	NotifyStatusChange(ctx context.Context, event sDto.StatusChangeResponse) error
	Subscribe(ctx context.Context) (<-chan sDto.StatusChangeResponse, error)
}
