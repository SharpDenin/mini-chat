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
	start := time.Now()
	defer func() {
		p.metrics.ObserveLatency("UpdateLastSeen", time.Since(start))
	}()

	if userId <= 0 {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidUserId,
			UserId: userId,
			Action: "UpdateLastSeen",
		}

		return middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	if err := p.validateUserExists(ctx, userId); err != nil {
		err = &sDto.PresenceError{
			Err:    sDto.ErrUserNotFound,
			UserId: userId,
			Action: "UpdateLastSeen",
		}

		return middleware_chat.NewCustomError(http.StatusNotFound, err.Error(), nil)
	}

	//// Проверяем rate limit (опционально)
	//if err := p.checkRateLimit(ctx, userId, "online"); err != nil {
	//	return err
	//}

	if err := p.repo.SetLastSeen(ctx, userId); err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"error":  err,
		}).Error("failed to update last seen")

		return middleware_chat.NewCustomError(http.StatusInternalServerError, "failed to update last seen", nil)
	}

	p.metrics.IncConnectionOpened(userId, "heartbeat")

	return nil
}

func (p *PresenceService) GetPresence(ctx context.Context, userId int64) (*sDto.PresenceResponse, error) {
	start := time.Now()
	defer func() {
		p.metrics.ObserveLatency("MarkOffline", time.Since(start))
	}()

	if userId <= 0 {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidUserId,
			UserId: userId,
			Action: "GetPresence",
		}

		return nil, middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	if err := p.validateUserExists(ctx, userId); err != nil {
		err = &sDto.PresenceError{
			Err:    sDto.ErrUserNotFound,
			UserId: userId,
			Action: "GetPresence",
		}

		return nil, middleware_chat.NewCustomError(http.StatusNotFound, err.Error(), nil)
	}

	repoPresence, err := p.repo.GetUserPresence(ctx, userId)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"error":  err,
		}).Error("failed to get presence")

		return nil, middleware_chat.NewCustomError(http.StatusInternalServerError, "failed to get presence", nil)
	}

	var presence *sDto.PresenceResponse
	if repoPresence != nil {
		presence = &sDto.PresenceResponse{
			UserId:   userId,
			LastSeen: repoPresence.LastSeen,
		}

		if repoPresence.Online {
			presence.Status = sDto.StatusOnline

			if time.Since(repoPresence.LastSeen) > p.config.IdleThreshold {
				presence.Status = sDto.StatusIdle
			}
		} else {
			presence.Status = sDto.StatusOffline
		}
	} else {
		presence = &sDto.PresenceResponse{
			UserId: userId,
			Status: sDto.StatusOffline,
		}
	}

	return presence, nil
}

func (p *PresenceService) GetBulkPresence(ctx context.Context, userIds []int64) (*sDto.BulkPresenceResponse, error) {
	start := time.Now()
	defer func() {
		p.metrics.ObserveLatency("GetBulkPresence", time.Since(start))
	}()

	if len(userIds) == 0 {
		return &sDto.BulkPresenceResponse{
			Presences: make(map[int64]sDto.PresenceResponse),
		}, nil
	}

	if len(userIds) > 1000 {
		userIds = userIds[:1000]
		p.log.WithFields(logrus.Fields{
			"requested":    len(userIds),
			"truncated_to": 1000,
		}).Warn("too many users")
	}

	response := &sDto.BulkPresenceResponse{
		Presences: make(map[int64]sDto.PresenceResponse),
		Errors:    make(map[int64]sDto.PresenceError),
	}

	for _, userId := range userIds {
		if userId <= 0 {
			err := &sDto.PresenceError{
				Err:    sDto.ErrInvalidUserId,
				UserId: userId,
				Action: "GetBulkPresence",
			}

			p.log.WithFields(logrus.Fields{
				"userId": userId,
				"error":  err,
			}).Warn("one of userId is invalid, it will be ignored")

			continue
		}
		if err := p.validateUserExists(ctx, userId); err != nil {
			err = &sDto.PresenceError{
				Err:    sDto.ErrUserNotFound,
				UserId: userId,
				Action: "GetBulkPresence",
			}

			p.log.WithFields(logrus.Fields{
				"userId": userId,
				"error":  err,
			}).Warn("one of user does not exist, it will be ignored")

			continue
		}

		presence, err := p.repo.GetUserPresence(ctx, userId)
		if err != nil {
			err := &sDto.PresenceError{
				Err:    sDto.ErrServiceUnavailable,
				UserId: userId,
				Action: "GetBulkPresence",
			}

			response.Errors[userId] = *err

			p.log.WithFields(logrus.Fields{
				"userId": userId,
				"error":  err,
			}).Warn("failed to get presence")

			continue
		}

		sPresence := &sDto.PresenceResponse{
			UserId:   presence.UserId,
			Status:   sDto.UserStatus(presence.Status),
			LastSeen: presence.LastSeen,
		}

		response.Presences[userId] = *sPresence
	}

	return response, nil
}

func (p *PresenceService) GetOnlineUsers(ctx context.Context, userIds []int64) ([]int64, error) {
	start := time.Now()
	defer func() {
		p.metrics.ObserveLatency("GetOnlineUsers", time.Since(start))
	}()

	if len(userIds) == 0 {
		return []int64{}, nil
	}

	checkedUserIds := make([]int64, 0, len(userIds))
	for _, userId := range userIds {
		if userId <= 0 {
			err := &sDto.PresenceError{
				Err:    sDto.ErrInvalidUserId,
				UserId: userId,
				Action: "GetBulkPresence",
			}

			p.log.WithFields(logrus.Fields{
				"userId": userId,
				"error":  err,
			}).Warn("one of userId is invalid, it will be ignored")

			break
		}
		if err := p.validateUserExists(ctx, userId); err != nil {
			err = &sDto.PresenceError{
				Err:    sDto.ErrUserNotFound,
				UserId: userId,
				Action: "GetBulkPresence",
			}

			p.log.WithFields(logrus.Fields{
				"userId": userId,
				"error":  err,
			}).Warn("one of user does not exist, it will be ignored")

			break
		}

		checkedUserIds = append(checkedUserIds, userId)
	}

	onlineUsers, err := p.repo.GetOnlineFriends(ctx, 0, checkedUserIds)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"error":      err,
			"user_count": len(userIds),
		}).Warn("failed to get online users")

		return nil, middleware_chat.NewCustomError(http.StatusInternalServerError, "Failed to get online users", err)
	}

	return onlineUsers, nil
}

func (p *PresenceService) GetRecentlyOnline(ctx context.Context, since time.Time) ([]int64, error) {
	//TODO implement me
	panic("implement me")
}

//TODO Реализовать (service + repo)
//func (p *PresenceService) AddConnection(ctx context.Context, userId int64, connId int64, deviceType string) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (p *PresenceService) RemoveConnection(ctx context.Context, userId int64, connId int64) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (p *PresenceService) GetUserConnections(ctx context.Context, userId int64) ([]int64, error) {
//	//TODO implement me
//	panic("implement me")
//}

func (p *PresenceService) CleanupStaleData(ctx context.Context) error {
	p.log.Info("starting stale data cleanup")

	if err := p.repo.CleanupStaleOnline(ctx); err != nil {
		p.log.Errorf("failed to cleanup stale data: %v", err)

		return middleware_chat.NewCustomError(http.StatusInternalServerError, "Failed to cleanup stale data", err)
	}

	p.log.Info("finished stale data cleanup")

	return nil
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
