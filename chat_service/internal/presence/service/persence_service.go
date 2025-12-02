package service

import (
	pRepo "chat_service/internal/presence/repository"
	sDto "chat_service/internal/presence/service/dto"
	"context"
	"log/slog"
	"time"
)

type PresenceService struct {
	repo     pRepo.PresenceRepoInterface
	logger   *slog.Logger
	config   *Config
	metrics  MetricsCollector
	notifier StatusNotifier
}

type Config struct {
	MaxConnectionsPerUser int
	RateLimitPerUser      time.Duration
	CleanupInterval       time.Duration
	IdleThreshold         time.Duration
}

type MetricsCollector interface {
	IncStatusChange(userId, from, to string)
	IncConnectionOpened(userId, op string)
	ObserveLatency(method string, duration time.Duration)
}

type StatusNotifier interface {
	NotifyStatusChange(ctx context.Context, event sDto.StatusChangeEvent) error
}

func NewPresenceService(repo pRepo.PresenceRepoInterface, logger *slog.Logger,
	config *Config, metrics MetricsCollector,
	notifier StatusNotifier) PresenceServiceInterface {
	if logger == nil {
		logger = slog.Default()
	}

	return &PresenceService{
		repo:     repo,
		logger:   logger,
		config:   config,
		metrics:  metrics,
		notifier: notifier,
	}
}

func (p PresenceService) MarkOnline(ctx context.Context, userId string, opt ...sDto.MarkOptionRequest) error {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) MarkOffline(ctx context.Context, userId string, opt ...sDto.MarkOptionRequest) error {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) UpdateLastSeen(ctx context.Context, userId string) error {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) GetPresence(ctx context.Context, userId string) (*sDto.PresenceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) GetBulkPresence(ctx context.Context, userIds []string) (*sDto.BulkPresenceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) GetOnlineUsers(ctx context.Context, userIds []string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) GetRecentlyOnline(ctx context.Context, since time.Time) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) AddConnection(ctx context.Context, userId string, connId string, deviceType string) error {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) RemoveConnection(ctx context.Context, userId string, connId string) error {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) GetUserConnections(ctx context.Context, userId string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) CleanupStaleData(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) HealthCheck(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (p PresenceService) SubscribeStatusChanges(ctx context.Context) (<-chan sDto.StatusChangeEvent, error) {
	//TODO implement me
	panic("implement me")
}
