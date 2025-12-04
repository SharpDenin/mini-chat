package repository

import (
	"chat_service/internal/presence/config"
	rModels "chat_service/internal/presence/repository/repo_dto"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	onlineUsersZSet = "presence:online"
	statusTTL       = 15 * time.Minute
)

type RedisRepo struct {
	client *redis.Client
	config *config.RedisConfig
}

func NewRedisRepo(config *config.RedisConfig) (PresenceRepoInterface, error) {
	db, err := strconv.Atoi(config.DB)
	if err != nil {
		return nil, fmt.Errorf("invalid DB number: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       db,

		PoolSize:     100,
		MinIdleConns: 10,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
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

func (r *RedisRepo) SetOnline(ctx context.Context, userId int64) error {
	now := time.Now().UnixMilli()
	userKey := r.userKey(userId)
	onlineSet := r.onlineSetKey()

	script := redis.NewScript(`
			redis.call('HSET', KEYS[1],
					'status', 'online',
					'lastSeen', ARGV[1]
			redis.call('EXPIRE', KEYS[1], ARGV[2])
			redis.call('ZADD', KEYS[2], ARGV[1], KEYS[3])
			
			return 1
	`)

	ttlSeconds := int(r.config.StatusTTL.Seconds())
	return script.Run(ctx, r.client,
		[]string{userKey, onlineSet, strconv.FormatInt(userId, 10)},
		ttlSeconds, now).Err()
}

func (r *RedisRepo) SetOffline(ctx context.Context, userId int64) error {
	userKey := r.userKey(userId)
	onlineSet := r.onlineSetKey()

	script := redis.NewScript(`
			redis.call('HSET', KEYS[1], 'status', 'offline')
			redis.call('ZREM', KEYS[2], KEYS[3])
			
			return 1
	`)

	return script.Run(ctx, r.client,
		[]string{userKey, onlineSet, strconv.FormatInt(userId, 10)}).Err()
}

func (r *RedisRepo) SetLastSeen(ctx context.Context, userId int64) error {
	now := time.Now().UnixMilli()
	userKey := r.userKey(userId)
	onlineSet := r.onlineSetKey()

	script := redis.NewScript(`
			local currentStatus = redis.call('HGET', KEYS[1], 'status')
			
			if currentStatus == 'online' then
					redis.call('HSET', KEYS[1], 'lastSeen', ARGV[1])
					redis.call('EXPIRE', KEYS[1], ARGV[2])
 					redis.call('ZADD', KEYS[2], ARGV[1], KEYS[3])
			end
	
			return 1
	`)

	ttlSeconds := int(r.config.StatusTTL.Seconds())
	return script.Run(ctx, r.client,
		[]string{userKey, onlineSet, strconv.FormatInt(userId, 10)},
		ttlSeconds, now).Err()
}

func (r *RedisRepo) IsOnline(ctx context.Context, userId int64) (bool, error) {
	onlineSet := r.onlineSetKey()

	score, err := r.client.ZScore(ctx, onlineSet, strconv.FormatInt(userId, 10)).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("redis zscore failed: %w", err)
	}

	lastSeenTime := time.UnixMilli(int64(score))
	timeSinceLastSeen := time.Since(lastSeenTime)

	if timeSinceLastSeen > r.config.StatusTTL {
		go func() {
			ctxBg := context.Background()
			r.client.ZRem(ctxBg, onlineSet, userId)
		}()
		return false, nil
	}

	return true, nil
}

func (r *RedisRepo) GetLastSeen(ctx context.Context, userId int64) (time.Time, error) {
	userKey := r.userKey(userId)

	tsStr, err := r.client.HGet(ctx, userKey, "lastSeen").Result()
	if errors.Is(err, redis.Nil) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get lastSeen: %w", err)
	}

	tsMs, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid lastSeen format: %w", err)
	}

	return time.UnixMilli(tsMs), nil
}

func (r *RedisRepo) GetOnlineFriends(ctx context.Context, userId int64, friendsIds []int64) ([]int64, error) {
	if len(friendsIds) == 0 {
		return []int64{}, nil
	}

	onlineSet := r.onlineSetKey()
	now := time.Now()
	threshold := now.Add(-r.config.StatusTTL).UnixMilli()

	cmds, err := r.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, friendId := range friendsIds {
			pipe.ZScore(ctx, onlineSet, strconv.FormatInt(friendId, 10))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("pipeline failed: %w", err)
	}

	onlineFriends := make([]int64, 0, len(friendsIds))

	for i, cmd := range cmds {
		score, err := cmd.(*redis.FloatCmd).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			continue
		}

		if int64(score) >= threshold {
			onlineFriends = append(onlineFriends, friendsIds[i])
		}
	}

	return onlineFriends, nil
}

func (r *RedisRepo) CleanupStaleOnline(ctx context.Context) error {
	onlineSet := r.onlineSetKey()
	threshold := float64(time.Now().Add(-r.config.StatusTTL).UnixMilli())

	return r.client.ZRemRangeByScore(ctx, onlineSet, "0", strconv.FormatFloat(threshold, 'f', -1, 64)).Err()
}

func (r *RedisRepo) GetUserPresence(ctx context.Context, userId int64) (*rModels.UserPresence, error) {
	userKey := r.userKey(userId)
	onlineSet := r.onlineSetKey()

	pipe := r.client.Pipeline()

	statusCmd := pipe.HGet(ctx, userKey, "status")
	lastSeenCmd := pipe.HGet(ctx, userKey, "lastSeen")

	onlineScoreCmd := pipe.ZScore(ctx, onlineSet, strconv.FormatInt(userId, 10))

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to get presence data: %w", err)
	}

	presence := &rModels.UserPresence{
		UserId: userId,
	}

	status, _ := statusCmd.Result()
	onlineScore, onlineErr := onlineScoreCmd.Result()
	if onlineErr == nil {
		if time.Since(time.UnixMilli(int64(onlineScore))) <= r.config.StatusTTL {
			presence.Online = true
			presence.Status = "online"
		} else {
			if status == "" {
				presence.Status = "offline"
			} else {
				presence.Status = status
			}
			presence.Online = false
		}
	} else {
		if status == "" {
			presence.Status = "offline"
		} else {
			presence.Status = status
		}
		presence.Online = false
	}

	if lastSeenStr, err := lastSeenCmd.Result(); err == nil && lastSeenStr != "" {
		if tsMs, err := strconv.ParseInt(lastSeenStr, 10, 64); err == nil {
			presence.LastSeen = time.UnixMilli(tsMs)
		}
	}

	return presence, nil
}

func (r *RedisRepo) userKey(userId int64) string {
	return fmt.Sprintf("user:presence:%v", userId)
}
func (r *RedisRepo) onlineSetKey() string {
	return fmt.Sprintf("%s:presence:online", r.config.Namespace)
}
