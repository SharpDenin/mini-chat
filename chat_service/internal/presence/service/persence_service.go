package service

import (
	"chat_service/internal/helpers"
	pRepo "chat_service/internal/presence/repository"
	sDto "chat_service/internal/presence/service/dto"
	"chat_service/middleware_chat"
	"chat_service/pkg/grpc_client"
	"chat_service/pkg/grpc_generated/profile"
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type PresenceService struct {
	profileClient *grpc_client.ProfileClient
	repo          pRepo.PresenceRepoInterface
	log           *logrus.Logger
	config        *Config
	metrics       MetricsCollector
	notifier      StatusNotifier
}

func NewPresenceService(repo pRepo.PresenceRepoInterface, log *logrus.Logger,
	config *Config, metrics MetricsCollector,
	notifier StatusNotifier) PresenceServiceInterface {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}

	return &PresenceService{
		repo:     repo,
		log:      log,
		config:   config,
		metrics:  metrics,
		notifier: notifier,
	}
}

func (p *PresenceService) MarkOnline(ctx context.Context, userId int64, opts ...sDto.MarkOptionRequest) error {
	start := time.Now()
	defer func() {
		p.metrics.ObserveLatency("MarkOnline", time.Since(start))
	}()

	if userId <= 0 {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidUserId,
			UserId: userId,
			Action: "MarkOnline",
		}

		return middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	if err := p.validateUserExists(ctx, userId); err != nil {
		err = &sDto.PresenceError{
			Err:    sDto.ErrUserNotFound,
			UserId: userId,
			Action: "MarkOnline",
		}

		return middleware_chat.NewCustomError(http.StatusNotFound, err.Error(), nil)
	}

	options := &sDto.MarkOptions{
		Source:    "api",
		Timestamp: time.Now(),
	}
	for _, o := range opts {
		o(options)
	}

	//// Проверяем rate limit (опционально)
	//if err := p.checkRateLimit(ctx, userId, "online"); err != nil {
	//	return err
	//}

	currentPresence, err := p.repo.GetUserPresence(ctx, userId)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"error":  err,
		}).Error("failed to get current presence")
	}

	if currentPresence != nil && currentPresence.Online && !options.Force {
		err = &sDto.PresenceError{
			Err:    sDto.ErrAlreadyOnline,
			UserId: userId,
			Action: "MarkOnline",
		}

		return middleware_chat.NewCustomError(http.StatusConflict, err.Error(), nil)
	}

	if err = p.repo.SetOnline(ctx, userId); err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"error":  err,
		}).Error("Failed to mark online")

		return middleware_chat.NewCustomError(http.StatusInternalServerError, "failed to mark online", nil)
	}

	if p.notifier != nil && currentPresence != nil {
		event := sDto.StatusChangeEvent{
			UserId:    userId,
			OldStatus: sDto.StatusOffline,
			NewStatus: sDto.StatusOnline,
			Timestamp: options.Timestamp,
			Source:    options.Source,
		}

		if err = p.notifier.NotifyStatusChange(ctx, event); err != nil {
			p.log.WithFields(logrus.Fields{
				"userId": userId,
				"error":  err,
			}).Error("failed to notify status change event")
		}
	}

	p.metrics.IncStatusChange(userId, "offline", "online")
	p.metrics.IncConnectionOpened(userId, "connect")

	p.log.WithFields(logrus.Fields{
		"userId":    userId,
		"source":    options.Source,
		"device_id": options.DeviceId,
	}).Info("marked online")

	return nil
}

func (p *PresenceService) MarkOffline(ctx context.Context, userId int64, opts ...sDto.MarkOptionRequest) error {
	start := time.Now()
	defer func() {
		p.metrics.ObserveLatency("MarkOffline", time.Since(start))
	}()

	if userId <= 0 {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidUserId,
			UserId: userId,
			Action: "MarkOffline",
		}

		return middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	if err := p.validateUserExists(ctx, userId); err != nil {
		err = &sDto.PresenceError{
			Err:    sDto.ErrUserNotFound,
			UserId: userId,
			Action: "MarkOnline",
		}

		return middleware_chat.NewCustomError(http.StatusNotFound, err.Error(), nil)
	}

	options := &sDto.MarkOptions{
		Source:    "api",
		Timestamp: time.Now(),
	}
	for _, o := range opts {
		o(options)
	}

	//// Проверяем rate limit (опционально)
	//if err := p.checkRateLimit(ctx, userId, "online"); err != nil {
	//	return err
	//}

	currentPresence, err := p.repo.GetUserPresence(ctx, userId)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"error":  err,
		}).Error("failed to get current presence")
	}

	if currentPresence != nil && !currentPresence.Online && !options.Force {
		err = &sDto.PresenceError{
			Err:    sDto.ErrAlreadyOffline,
			UserId: userId,
			Action: "MarkOffline",
		}

		return middleware_chat.NewCustomError(http.StatusConflict, err.Error(), nil)
	}

	if err = p.repo.SetOffline(ctx, userId); err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"error":  err,
		}).Error("failed to mark online")

		return middleware_chat.NewCustomError(http.StatusInternalServerError, "failed to mark online", nil)
	}

	if p.notifier != nil && currentPresence != nil && currentPresence.Online {
		event := sDto.StatusChangeEvent{
			UserId:    userId,
			OldStatus: sDto.StatusOnline,
			NewStatus: sDto.StatusOffline,
			Timestamp: options.Timestamp,
			Source:    options.Source,
		}

		if err = p.notifier.NotifyStatusChange(ctx, event); err != nil {
			p.log.WithFields(logrus.Fields{
				"userId": userId,
				"error":  err,
			}).Error("failed to notify status change event")
		}
	}

	p.metrics.IncStatusChange(userId, "online", "offline")
	p.metrics.IncConnectionOpened(userId, "disconnect")

	p.log.WithFields(logrus.Fields{
		"userId": userId,
		"source": options.Source,
	}).Info("marked offline")

	return nil
}

func (p *PresenceService) UpdateLastSeen(ctx context.Context, userId int64) error {
	//TODO implement me
	panic("implement me")
}

func (p *PresenceService) GetPresence(ctx context.Context, userId int64) (*sDto.PresenceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PresenceService) GetBulkPresence(ctx context.Context, userIds []int64) (*sDto.BulkPresenceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PresenceService) GetOnlineUsers(ctx context.Context, userIds []int64) ([]int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PresenceService) GetRecentlyOnline(ctx context.Context, since time.Time) ([]int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PresenceService) AddConnection(ctx context.Context, userId int64, connId int64, deviceType string) error {
	//TODO implement me
	panic("implement me")
}

func (p *PresenceService) RemoveConnection(ctx context.Context, userId int64, connId int64) error {
	//TODO implement me
	panic("implement me")
}

func (p *PresenceService) GetUserConnections(ctx context.Context, userId int64) ([]int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PresenceService) CleanupStaleData(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (p *PresenceService) HealthCheck(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (p *PresenceService) SubscribeStatusChanges(ctx context.Context) (<-chan sDto.StatusChangeEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PresenceService) validateUserExists(ctx context.Context, userId int64) error {
	userReq := &profile.UserExistsRequest{UserId: strconv.FormatInt(userId, 10)}
	exist, err := helpers.CheckUserExist(ctx, p.profileClient, userReq)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"user_id": userId,
			"error":   err,
		}).Warn("Failed to get user")
		return errors.New("failed to get user")
	}
	if !exist {
		p.log.WithField("user_id", userId).Warn("User not found")
		return errors.New("user not found")
	}
	return nil
}
