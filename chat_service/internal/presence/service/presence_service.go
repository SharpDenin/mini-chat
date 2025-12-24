package service

import (
	"chat_service/internal/presence/config"
	pRepo "chat_service/internal/presence/repository"
	sDto "chat_service/internal/presence/service/dto"
	"chat_service/middleware_chat"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type PresenceService struct {
	repo     pRepo.PresenceRepoInterface
	log      *logrus.Logger
	config   *config.RedisServiceConfig
	metrics  MetricsCollector
	notifier StatusNotifier
}

func NewPresenceService(repo pRepo.PresenceRepoInterface, log *logrus.Logger,
	config *config.RedisServiceConfig, metrics MetricsCollector,
	notifier StatusNotifier) PresenceServiceInterface {
	if metrics == nil {
		metrics = &nullMetricsCollector{}
	}
	if notifier == nil {
		notifier = &nullStatusNotifier{}
	}
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
		event := sDto.StatusChangeResponse{
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
		event := sDto.StatusChangeResponse{
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
		p.metrics.ObserveLatency("GetPresence", time.Since(start))
	}()

	if userId <= 0 {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidUserId,
			UserId: userId,
			Action: "GetPresence",
		}

		return nil, middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	repoPresence, err := p.repo.GetUserPresence(ctx, userId)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"error":  err,
		}).Error("failed to get presence")

		return &sDto.PresenceResponse{
			UserId:   userId,
			Status:   sDto.StatusOffline,
			LastSeen: time.Time{},
		}, nil
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
			UserId:   userId,
			Status:   sDto.StatusOffline,
			LastSeen: time.Time{},
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
			Presences: make(map[int64]*sDto.PresenceResponse),
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
		Presences: make(map[int64]*sDto.PresenceResponse),
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

		response.Presences[userId] = sPresence
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

			continue
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
	start := time.Now()
	defer func() {
		p.metrics.ObserveLatency("GetRecentlyOnline", time.Since(start))
	}()

	if since.IsZero() {
		since = time.Now().Add(-24 * time.Hour)
	}

	if since.After(time.Now()) {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidTimestamp,
			UserId: 0,
			Action: "GetBulkPresence",
		}

		return nil, middleware_chat.NewCustomError(http.StatusInternalServerError, "Future timestamp not allowed", err)
	}

	recentlyOnlineUsers, err := p.repo.GetRecentlyOnline(ctx, since)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"since": since,
			"error": err,
		}).Error("failed to get recently online users")

		return nil, middleware_chat.NewCustomError(http.StatusInternalServerError, "Failed to get recently online users", err)
	}

	p.log.WithFields(logrus.Fields{
		"since":       since,
		"users_count": len(recentlyOnlineUsers),
		"duration_ms": time.Since(start).Milliseconds(),
	}).Info("get recently online users")

	return recentlyOnlineUsers, nil
}

func (p *PresenceService) AddConnection(ctx context.Context, userId int64, connId int64, deviceType string) error {
	start := time.Now()
	defer func() {
		p.metrics.ObserveLatency("AddConnection", time.Since(start))
	}()

	if userId <= 0 {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidUserId,
			UserId: userId,
			Action: "AddConnection",
		}

		return middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	if connId <= 0 {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidConnId,
			UserId: userId,
			Action: "AddConnection",
		}

		return middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	validDevices := map[string]bool{
		sDto.DeviceWeb:     true,
		sDto.DeviceDesktop: true,
		sDto.DeviceAndroid: true,
		sDto.DeviceIOS:     true,
	}

	if !validDevices[deviceType] {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidDeviceType,
			UserId: userId,
			Action: "AddConnection",
		}

		return middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	existingConn, err := p.repo.GetConnectionInfo(ctx, userId, connId)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"connId": connId,
			"error":  err,
		}).Warn("failed to check existing conn")
	}

	if existingConn != nil {
		if err := p.repo.UpdateConnectionActivity(ctx, userId, connId); err != nil {
			p.log.WithFields(logrus.Fields{
				"userId": userId,
				"connId": connId,
				"error":  err,
			}).Error("failed to update connection activity")

			return middleware_chat.NewCustomError(http.StatusInternalServerError, "Failed to update connection activity", nil)
		}

		p.metrics.IncConnectionOpened(userId, "reconnect")

	} else {
		if err := p.repo.AddConnection(ctx, userId, connId, deviceType); err != nil {
			p.log.WithFields(logrus.Fields{
				"userId":     userId,
				"connId":     connId,
				"deviceType": deviceType,
				"error":      err,
			}).Error("failed to add connection")

			return middleware_chat.NewCustomError(http.StatusInternalServerError, "Failed to add connection", nil)
		}

		p.metrics.IncConnectionOpened(userId, "new_connection")

	}

	p.log.WithFields(logrus.Fields{
		"userId":     userId,
		"connId":     connId,
		"deviceType": deviceType,
	}).Info("added/updated connection")

	return nil
}

func (p *PresenceService) RemoveConnection(ctx context.Context, userId int64, connId int64) error {
	start := time.Now()
	defer func() {
		p.metrics.ObserveLatency("RemoveConnection", time.Since(start))
	}()

	if userId <= 0 {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidUserId,
			UserId: userId,
			Action: "RemoveConnection",
		}

		return middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	if connId <= 0 {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidConnId,
			UserId: userId,
			Action: "RemoveConnection",
		}

		return middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	existingConn, err := p.repo.GetConnectionInfo(ctx, userId, connId)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"connId": connId,
			"error":  err,
		}).Warn("failed to get connection info")
	}

	if existingConn == nil {
		err := &sDto.PresenceError{
			Err:    sDto.ErrConnectionNotFound,
			UserId: userId,
			Action: "RemoveConnection",
		}

		return middleware_chat.NewCustomError(http.StatusInternalServerError, err.Error(), nil)
	}

	if err := p.repo.RemoveConnection(ctx, userId, connId); err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"connId": connId,
			"error":  err,
		}).Error("failed to remove connection")

		return middleware_chat.NewCustomError(http.StatusInternalServerError, err.Error(), nil)
	}

	remainingConns, err := p.repo.GetUserConnections(ctx, userId)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"error":  err,
		}).Warn("failed to get remaining connections")
	}

	if len(remainingConns) == 0 {
		if err := p.MarkOffline(ctx, userId, sDto.WithSource("connection_removed")); err != nil {
			p.log.WithFields(logrus.Fields{
				"userId": userId,
				"error":  err,
			}).Error("failed to mark user offline after last connection removed")
		}
	}

	p.metrics.IncConnectionOpened(userId, "connection_removed")

	p.log.WithFields(logrus.Fields{
		"userId": userId,
		"connId": connId,
	}).Info("removed connection")

	return nil
}

func (p *PresenceService) GetUserConnections(ctx context.Context, userId int64) ([]*sDto.ConnectionInfoResponse, error) {
	start := time.Now()
	defer func() {
		p.metrics.ObserveLatency("RemoveConnection", time.Since(start))
	}()

	if userId <= 0 {
		err := &sDto.PresenceError{
			Err:    sDto.ErrInvalidUserId,
			UserId: userId,
			Action: "RemoveConnection",
		}

		return nil, middleware_chat.NewCustomError(http.StatusBadRequest, err.Error(), nil)
	}

	repoConnections, err := p.repo.GetAllUserConnections(ctx, userId)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"error":  err,
		}).Error("failed to get user connections")

		return nil, middleware_chat.NewCustomError(http.StatusInternalServerError, "Failed to get user connections", nil)
	}

	conns := make([]*sDto.ConnectionInfoResponse, 0, len(repoConnections))
	for _, repoConn := range repoConnections {
		conn := &sDto.ConnectionInfoResponse{
			ConnId:       repoConn.ConnId,
			UserId:       repoConn.UserId,
			DeviceType:   repoConn.DeviceType,
			ConnectedAt:  repoConn.ConnectedAt,
			LastActivity: repoConn.LastActivity,
		}

		conns = append(conns, conn)
	}

	return conns, nil
}

func (p *PresenceService) UpdateConnectionActivity(ctx context.Context, userId int64, connId int64) error {
	if userId <= 0 || connId <= 0 {
		return nil
	}

	if err := p.repo.UpdateConnectionActivity(ctx, userId, connId); err != nil {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"connId": connId,
			"error":  err,
		}).Warn("failed to update connection activity")
	}

	return nil
}

func (p *PresenceService) GetConnectionStats(ctx context.Context, userId int64) (*sDto.ConnectionStatsResponse, error) {
	connections, err := p.repo.GetAllUserConnections(ctx, userId)
	if err != nil {
		return nil, err
	}

	stats := &sDto.ConnectionStatsResponse{
		TotalConnections: len(connections),
		Devices:          make(map[string]int64),
	}

	for _, conn := range connections {
		stats.Devices[conn.DeviceType]++
	}

	return stats, nil
}

func (p *PresenceService) CleanupStaleData(ctx context.Context) error {
	p.log.Info("starting stale data cleanup")

	if err := p.repo.CleanupStaleOnline(ctx); err != nil {
		p.log.Errorf("failed to cleanup stale data: %v", err)

		return middleware_chat.NewCustomError(http.StatusInternalServerError, "Failed to cleanup stale data", err)
	}

	p.log.Info("finished stale data cleanup")

	return nil
}

func (p *PresenceService) SubscribeStatusChanges(ctx context.Context) (<-chan sDto.StatusChangeResponse, error) {
	if p.notifier == nil {
		return nil, middleware_chat.NewCustomError(http.StatusServiceUnavailable, "status change notifications not available", nil)
	}

	ch, err := p.notifier.Subscribe(ctx)
	if err != nil {
		p.log.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to subscribe to status change notifications")

		return nil, middleware_chat.NewCustomError(http.StatusInternalServerError, "Failed to subscribe to status change notifications", nil)
	}

	return ch, nil
}
