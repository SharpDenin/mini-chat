package presence_repository

import (
	"chat_service/internal/presence/presence_config"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	onlineUsersZSet = "presence:online"
)

type RedisRepo struct {
	client *redis.Client
	config *presence_config.RedisConfig
}

func NewRedisRepo(config *presence_config.RedisConfig) (PresenceRepo, error) {
	db, err := strconv.Atoi(config.DB)
	if err != nil {
		return nil, fmt.Errorf("invalid DB number: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisRepo{
		client: client,
		config: config,
	}, nil
}

func (r *RedisRepo) SetOnline(ctx context.Context, userId string) error {
	now := time.Now().UnixMilli()
	key := r.userKey(userId)

	pipe := r.client.Pipeline()

	pipe.HSet(ctx, key, map[string]interface{}{
		"status":   "online",
		"lastSeen": now,
	})

	pipe.ZAdd(ctx, onlineUsersZSet, redis.Z{
		Score:  float64(now),
		Member: userId,
	})

	pipe.Expire(ctx, key, 15*time.Minute)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisRepo) SetOffline(ctx context.Context, userId string) error {
	key := r.userKey(userId)

	pipe := r.client.Pipeline()
	pipe.HSet(ctx, key, "status", "offline")
	pipe.ZRem(ctx, onlineUsersZSet, userId)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisRepo) SetLastSeen(ctx context.Context, userId string) error {
	now := time.Now().UnixMilli()
	key := r.userKey(userId)

	pipe := r.client.Pipeline()

	pipe.HIncrBy(ctx, key, "lastSeen", now)
	pipe.HSet(ctx, key, "lastSeen", now)

	status, err := r.client.HGet(ctx, key, "status").Result()
	if errors.Is(err, redis.Nil) {
		pipe.HSet(ctx, key, "status", "offline")
	} else if err != nil {
		return err
	} else if status != "online" {
		pipe.ZAdd(ctx, onlineUsersZSet, redis.Z{
			Score:  float64(now),
			Member: userId,
		})
	}

	pipe.Expire(ctx, key, 15*time.Minute)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisRepo) IsOnline(ctx context.Context, userId string) (bool, error) {
	score, err := r.client.ZScore(ctx, onlineUsersZSet, userId).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if time.Since(time.UnixMilli(int64(score))) > 15*time.Minute {
		r.client.ZRem(ctx, onlineUsersZSet, userId)
		return false, nil
	}

	return true, nil
}

func (r *RedisRepo) GetLastSeen(ctx context.Context, userId string) (time.Time, error) {
	tsStr, err := r.client.HGet(ctx, r.userKey(userId), "lastSeen").Result()
	if errors.Is(err, redis.Nil) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}

	tsMs, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.UnixMilli(tsMs), nil
}

func (r *RedisRepo) GetOnlineFriends(ctx context.Context, userId string, friendsIds []string) ([]string, error) {
	if len(friendsIds) == 0 {
		return []string{}, nil
	}

	cmds, err := r.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, id := range friendsIds {
			pipe.ZScore(ctx, onlineUsersZSet, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	online := make([]string, 0, len(cmds))
	now := time.Now()
	for i, cmd := range cmds {
		score, err := cmd.(*redis.FloatCmd).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			return nil, err
		}

		if now.Sub(time.UnixMilli(int64(score))) <= 15*time.Minute {
			online = append(online, friendsIds[i])
		}
	}

	return online, nil
}

func (r *RedisRepo) CleanupStaleOnline(ctx context.Context) error {
	threshold := float64(time.Now().Add(-15 * time.Minute).UnixMilli())
	return r.client.ZRemRangeByScore(ctx, onlineUsersZSet, "0", strconv.FormatFloat(threshold, 'f', -1, 64)).Err()
}

func (r *RedisRepo) userKey(userId string) string {
	return fmt.Sprintf("user:presence:%s", userId)
}
