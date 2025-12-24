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

type nullMetricsCollector struct{}

func (n *nullMetricsCollector) ObserveLatency(method string, duration time.Duration) {
	// Ничего не делаем
}

func (n *nullMetricsCollector) IncStatusChange(userId int64, from, to string) {
	// Ничего не делаем
}

func (n *nullMetricsCollector) IncConnectionOpened(userId int64, op string) {
	// Ничего не делаем
}

type StatusNotifier interface {
	NotifyStatusChange(ctx context.Context, event sDto.StatusChangeResponse) error
	Subscribe(ctx context.Context) (<-chan sDto.StatusChangeResponse, error)
}

type nullStatusNotifier struct{}

func (n *nullStatusNotifier) NotifyStatusChange(ctx context.Context, event sDto.StatusChangeResponse) error {
	// Ничего не делаем, возвращаем nil
	return nil
}

func (n *nullStatusNotifier) Subscribe(ctx context.Context) (<-chan sDto.StatusChangeResponse, error) {
	// Создаем пустой канал
	ch := make(chan sDto.StatusChangeResponse)
	close(ch) // Закрываем сразу, так как ничего не будем отправлять
	return ch, nil
}
