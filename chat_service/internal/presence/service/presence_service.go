package service

import (
	"chat_service/internal/presence/config"
	"chat_service/internal/presence/repository"
	rDto "chat_service/internal/presence/repository/repo_dto"
	sDto "chat_service/internal/presence/service/dto"
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type presenceService struct {
	repo          repository.PresenceRepo
	idleThreshold time.Duration
}

func NewPresenceService(
	repo repository.PresenceRepo,
	cfg *config.RedisConfig,
) PresenceService {
	return &presenceService{
		repo:          repo,
		idleThreshold: cfg.IdleThreshold,
	}
}

func (s *presenceService) OnConnect(ctx context.Context, userId, connId int64, device string) error {
	return s.repo.AddConnection(ctx, userId, connId, device)
}

func (s *presenceService) OnDisconnect(ctx context.Context, userId, connId int64) error {
	return s.repo.RemoveConnection(ctx, userId, connId)
}

func (s *presenceService) OnHeartbeat(ctx context.Context, connId int64) error {
	err := s.repo.TouchConnection(ctx, connId)
	if err != nil {
		if err == redis.Nil {
			// heartbeat от зомби соединения
			log.Printf("[presence] heartbeat ignored, conn %d not found", connId)
			return nil
		}
		return err
	}

	return nil
}

func (s *presenceService) GetPresence(ctx context.Context, userId int64) *sDto.Presence {
	conns, err := s.repo.GetUserConnections(ctx, userId)
	if err != nil {
		// fail-safe: считаем offline
		return &sDto.Presence{
			UserId: userId,
			Status: sDto.Offline,
		}
	}

	if len(conns) == 0 {
		return &sDto.Presence{
			UserId: userId,
			Status: sDto.Offline,
		}
	}

	lastActivity := maxLastActivity(conns)

	status := sDto.Online
	if time.Since(lastActivity) > s.idleThreshold {
		status = sDto.Idle
	}

	return &sDto.Presence{
		UserId:   userId,
		Status:   status,
		LastSeen: lastActivity,
	}
}

func (s *presenceService) GetOnlineFriends(ctx context.Context, userId int64, friends []int64) []int64 {
	if len(friends) == 0 {
		return nil
	}

	online := make([]int64, 0, len(friends))

	for _, friendId := range friends {
		p := s.GetPresence(ctx, friendId)
		if p.Status == sDto.Online || p.Status == sDto.Idle {
			online = append(online, friendId)
		}
	}

	return online
}

func maxLastActivity(conns []rDto.Connection) time.Time {
	var maxLA time.Time

	for _, c := range conns {
		if c.LastActivity.After(maxLA) {
			maxLA = c.LastActivity
		}
	}

	return maxLA
}
