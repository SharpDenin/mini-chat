package repository

import (
	"chat_service/internal/presence/repository/repo_dto"
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed scripts/addConnection.lua
var addConnectionLua string

//go:embed scripts/touchConnection.lua
var touchConnectionLua string

//go:embed scripts/removeConnection.lua
var removeConnectionLua string

//go:embed scripts/cleanupDanglingConnections.lua
var cleanupDanglingConnectionsLua string

type redisPresenceRepo struct {
	rdb *redis.Client
	ttl time.Duration

	addConnScript     *redis.Script
	touchConnScript   *redis.Script
	removeConnScript  *redis.Script
	cleanupConnScript *redis.Script
}

func NewPresenceRepo(
	rdb *redis.Client,
	connectionTTL time.Duration,
) PresenceRepo {
	return &redisPresenceRepo{
		rdb: rdb,
		ttl: connectionTTL,

		addConnScript:     redis.NewScript(addConnectionLua),
		touchConnScript:   redis.NewScript(touchConnectionLua),
		removeConnScript:  redis.NewScript(removeConnectionLua),
		cleanupConnScript: redis.NewScript(cleanupDanglingConnectionsLua),
	}
}

func (r *redisPresenceRepo) AddConnection(ctx context.Context, userId, connId int64, device string) error {
	now := time.Now().UnixMilli()

	_, err := r.addConnScript.Run(ctx, r.rdb,
		[]string{
			connKey(connId),
			userConnSetKey(userId),
		}, userId, connId, device, now,
		int(r.ttl.Seconds()),
	).Result()

	return err
}

func (r *redisPresenceRepo) RemoveConnection(ctx context.Context, userId, connId int64) error {
	_, err := r.removeConnScript.Run(ctx, r.rdb,
		[]string{
			connKey(connId),
			userConnSetKey(userId),
		}, connId,
	).Result()

	return err
}

func (r *redisPresenceRepo) TouchConnection(ctx context.Context, connId int64) error {
	now := time.Now().UnixMilli()

	res, err := r.touchConnScript.Run(ctx, r.rdb, []string{connKey(connId)}, now, int(r.ttl.Seconds())).Int()

	if err != nil {
		return err
	}

	if res == 0 {
		return redis.Nil
	}

	return nil
}

func (r *redisPresenceRepo) GetUserConnections(ctx context.Context, userId int64) ([]repo_dto.Connection, error) {
	connIds, err := r.rdb.SMembers(ctx, userConnSetKey(userId)).Result()
	if err != nil {
		return nil, err
	}

	if len(connIds) == 0 {
		return nil, nil
	}

	var result []repo_dto.Connection

	for _, connIdStr := range connIds {
		connId, err := strconv.ParseInt(connIdStr, 10, 64)
		if err != nil {
			continue
		}

		data, err := r.rdb.HGetAll(ctx, connKey(connId)).Result()
		if err != nil || len(data) == 0 {
			continue
		}

		connectedAtMs, _ := strconv.ParseInt(data["connected_at"], 10, 64)
		lastActivityMs, _ := strconv.ParseInt(data["last_activity"], 10, 64)
		userIdParsed, _ := strconv.ParseInt(data["user_id"], 10, 64)

		result = append(result, repo_dto.Connection{
			ConnId:       connId,
			UserId:       userIdParsed,
			Device:       data["device"],
			ConnectedAt:  time.UnixMilli(connectedAtMs),
			LastActivity: time.UnixMilli(lastActivityMs),
		})
	}

	return result, nil
}

func (r *redisPresenceRepo) CleanupDanglingConnections(ctx context.Context, userId int64) error {
	_, err := r.cleanupConnScript.Run(ctx, r.rdb, []string{userConnSetKey(userId)}).Result()

	return err
}

func connKey(connId int64) string {
	return fmt.Sprintf("conn:%d", connId)
}

func userConnSetKey(userId int64) string {
	return fmt.Sprintf("user:%d:conns", userId)
}
