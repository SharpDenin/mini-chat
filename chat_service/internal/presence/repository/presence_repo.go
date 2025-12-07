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

func (r *RedisRepo) GetUserPresence(ctx context.Context, userId int64) (*rModels.UserPresenceResponse, error) {
	userKey := r.userKey(userId)
	onlineSet := r.onlineSetKey()

	pipe := r.client.Pipeline()

	statusCmd := pipe.HGet(ctx, userKey, "status")
	lastSeenCmd := pipe.HGet(ctx, userKey, "lastSeen")

	onlineScoreCmd := pipe.ZScore(ctx, onlineSet, strconv.FormatInt(userId, 10))

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to get presence data: %w", err)
	}

	presence := &rModels.UserPresenceResponse{
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

func (r *RedisRepo) GetRecentlyOnline(ctx context.Context, since time.Time) ([]int64, error) {
	onlineSet := r.onlineSetKey()

	threshold := float64(since.UnixMilli())

	userIdsStr, err := r.client.ZRangeByScore(ctx, onlineSet, &redis.ZRangeBy{
		Min: strconv.FormatFloat(threshold, 'f', -1, 64),
		Max: "+inf",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get recently online users: %w", err)
	}

	userIds := make([]int64, 0, len(userIdsStr))
	for _, userIdStr := range userIdsStr {
		if userId, err := strconv.ParseInt(userIdStr, 10, 64); err == nil {
			userIds = append(userIds, userId)
		}
	}

	return userIds, nil
}

func (r *RedisRepo) AddConnection(ctx context.Context, userId int64, connId int64, deviceType string) error {
	now := time.Now().UnixMilli()
	connKey := r.connectionKey(userId, connId)
	userConnSet := r.userConnectionsSetKey(userId)
	deviceConnSet := r.deviceConnectionsKey(userId, deviceType)

	script := redis.NewScript(`
			local connKey = KEYS[1]
			local userConnSet = KEYS[2]
			local deviceConnSet = KEYS[3]

			local timestamp = ARGV[1]
			local deviceType = ARGV[2]
			local ttl = ARGV[3]
			
			redis.call('HSET', connKey,
					'user_id', KEYS[4],
					'conn_id', KEYS[5],
					'device_type', deviceType,
					'connected_at', timestamp,
					'last_activity', timestamp,
			)

			redis.call('EXPIRE', connKey, ttl)
			redis.call('SADD', userConnSet, KEYS[5])
			redis.call('SADD', deviceConnSet, KEYS[5])
			redis.call('EXPIRE', userConnSet, ttl)
			redis.call('EXPIRE', deviceConnSet, ttl)

			return 1
	`)

	ttlSeconds := int(r.config.StatusTTL.Seconds())

	return script.Run(ctx, r.client,
		[]string{
			connKey,
			userConnSet,
			deviceConnSet,
			strconv.FormatInt(userId, 10),
			strconv.FormatInt(connId, 10),
		},
		now,
		deviceType,
		ttlSeconds,
	).Err()
}

func (r *RedisRepo) RemoveConnection(ctx context.Context, userId int64, connId int64) error {
	connKey := r.connectionKey(userId, connId)
	userConnSet := r.userConnectionsSetKey(userId)

	connInfo, err := r.GetConnectionInfo(ctx, userId, connId)
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("failed to get connection info: %w", err)
	}

	script := redis.NewScript(`
			local connKey = KEYS[1]
			local userConnSet = KEYS[2]
			local deviceConnSet = KEYS[3]
			
			redis.call('DEL', connKey)
			redis.call('SREM', userConnSet, KEYS[4)
			
			if deviceConnSet ~= '' then
					redis.call('SREM', deviceConnSet, KEYS[4])
			end
			
			local count = redis.call('SCARD', userConnSet)
			if count == 0 then
					redis.call('DEL', userConnSet)
			end
			
			if deviceConnSet ~= '' then
					local deviceCount = redis.call('SCARD', deviceConnSet)
					if deviceCount == 0 then
							redis.call('DEL', deviceConnSet)
					end
			end
			
			return 1
	`)

	deviceConnSet := ""
	if connInfo != nil {
		deviceConnSet = r.deviceConnectionsKey(userId, connInfo.DeviceType)
	}

	return script.Run(ctx, r.client,
		[]string{
			connKey,
			userConnSet,
			deviceConnSet,
			strconv.FormatInt(connId, 10),
		}).Err()
}

func (r *RedisRepo) GetUserConnections(ctx context.Context, userId int64) ([]int64, error) {
	userConnSet := r.userConnectionsSetKey(userId)

	connIdsStr, err := r.client.SMembers(ctx, userConnSet).Result()
	if errors.Is(err, redis.Nil) {
		return []int64{}, nil
	}
	if err != nil {
		return []int64{}, fmt.Errorf("failed to get user connections: %w", err)
	}

	connIds := make([]int64, 0, len(connIdsStr))
	for _, connIdStr := range connIdsStr {
		if connId, err := strconv.ParseInt(connIdStr, 10, 64); err == nil {
			connIds = append(connIds, connId)
		}
	}

	return connIds, nil
}

func (r *RedisRepo) GetConnectionInfo(ctx context.Context, userId int64, connId int64) (*rModels.ConnectionInfoResponse, error) {
	connKey := r.connectionKey(userId, connId)

	result, err := r.client.HGetAll(ctx, connKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get connection info: %w", err)
	}

	if len(result) == 0 {
		return nil, nil
	}

	connInfo := &rModels.ConnectionInfoResponse{
		UserId:     userId,
		ConnId:     connId,
		DeviceType: result["device_type"],
	}

	if connectedAtStr, ok := result["connected_at"]; ok && connectedAtStr != "" {
		if ts, err := strconv.ParseInt(connectedAtStr, 10, 64); err == nil {
			connInfo.ConnectedAt = time.UnixMilli(ts)
		}
	}

	if lastActivityStr, ok := result["last_activity"]; ok && lastActivityStr != "" {
		if ts, err := strconv.ParseInt(lastActivityStr, 10, 64); err == nil {
			connInfo.LastActivity = time.UnixMilli(ts)
		}
	}

	return connInfo, nil
}

func (r *RedisRepo) UpdateConnectionActivity(ctx context.Context, userId int64, connId int64) error {
	now := time.Now().UnixMilli()
	connKey := r.connectionKey(userId, connId)

	script := redis.NewScript(`
			local connKey = KEYS[1]
			local timestamp = ARGV[1]
			local ttl = ARGV[2]
			
			if redis.call('EXISTS', connKey) == 1 then
					redis.call('HSET', connKey, 'last_activity' timestamp)
					redis.call('EXPIRE', connKey, ttl)
					
					return 1
			end
			
			return 0
	`)

	ttlSeconds := int(r.config.StatusTTL.Seconds())

	return script.Run(ctx, r.client,
		[]string{connKey},
		now,
		ttlSeconds,
	).Err()
}

func (r *RedisRepo) GetAllUserConnections(ctx context.Context, userId int64) ([]rModels.ConnectionInfoResponse, error) {
	connIds, err := r.GetUserConnections(ctx, userId)
	if err != nil {
		return nil, err
	}

	pipe := r.client.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, 0, len(connIds))

	for i, connId := range connIds {
		cmds[i] = pipe.HGetAll(ctx, r.connectionKey(userId, connId))
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("pipeline failed (get all connections): %w", err)
	}

	connections := make([]rModels.ConnectionInfoResponse, 0, len(connIds))
	for i, connId := range connIds {
		result, err := cmds[i].Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			continue
		}

		if len(result) == 0 {
			continue
		}

		connInfo := &rModels.ConnectionInfoResponse{
			UserId:     userId,
			ConnId:     connId,
			DeviceType: result["device_type"],
		}

		if connectedAtStr, ok := result["connected_at"]; ok && connectedAtStr != "" {
			if ts, err := strconv.ParseInt(connectedAtStr, 10, 64); err == nil {
				connInfo.ConnectedAt = time.UnixMilli(ts)
			}
		}

		if lastActivityStr, ok := result["last_activity"]; ok && lastActivityStr != "" {
			if ts, err := strconv.ParseInt(lastActivityStr, 10, 64); err == nil {
				connInfo.LastActivity = time.UnixMilli(ts)
			}
		}

		connections = append(connections, *connInfo)
	}

	return connections, nil
}

func (r *RedisRepo) userKey(userId int64) string {
	return fmt.Sprintf("user:presence:%v", userId)
}

func (r *RedisRepo) onlineSetKey() string {
	return fmt.Sprintf("%s:presence:online", r.config.Namespace)
}

func (r *RedisRepo) connectionKey(userId int64, connId int64) string {
	return fmt.Sprintf("user:%d:conn:%d", userId, connId)
}

func (r *RedisRepo) userConnectionsSetKey(userId int64) string {
	return fmt.Sprintf("user:%d:connections", userId)
}

func (r *RedisRepo) deviceConnectionsKey(userId int64, deviceType string) string {
	return fmt.Sprintf("user:%d:device:%s:connections", userId, deviceType)
}
