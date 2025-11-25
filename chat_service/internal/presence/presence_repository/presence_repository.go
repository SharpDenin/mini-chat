package presence_repository

import (
	"chat_service/internal/presence/presence_config"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
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

func (r RedisRepo) SetOnline(ctx context.Context, userId string) error {
	//TODO implement me
	panic("implement me")
}

func (r RedisRepo) SetOffline(ctx context.Context, userId string) error {
	//TODO implement me
	panic("implement me")
}

func (r RedisRepo) SetLastSeen(ctx context.Context, userId string) error {
	//TODO implement me
	panic("implement me")
}

func (r RedisRepo) IsOnline(ctx context.Context, userId string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (r RedisRepo) GetLastSeen(ctx context.Context, userId string) (time.Time, error) {
	//TODO implement me
	panic("implement me")
}
