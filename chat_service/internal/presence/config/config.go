package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type RedisRepositoryConfig struct {
	Addr              string
	Password          string
	DB                string
	StatusTTL         time.Duration
	HeartbeatInterval time.Duration
	Namespace         string
}

type RedisServiceConfig struct {
	MaxConnectionsPerUser string
	RateLimitPerUser      time.Duration
	CleanupInterval       time.Duration
	IdleThreshold         time.Duration
}

func PRCfgLoad() (*RedisRepositoryConfig, error) {
	if err := godotenv.Load(".env"); err != nil {
		logrus.Error("Failed to load .env file: ", err)
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	toDuration := func(key string) time.Duration {
		secs, _ := strconv.Atoi(os.Getenv(key))
		return time.Duration(secs) * time.Second
	}

	config := &RedisRepositoryConfig{
		Addr:              os.Getenv("REDIS_ADDR"),
		Password:          os.Getenv("REDIS_PASSWORD"),
		DB:                os.Getenv("REDIS_DB"),
		StatusTTL:         toDuration("REDIS_STATUS_TTL"),
		HeartbeatInterval: toDuration("REDIS_HEARTBEAT_INTERVAL"),
		Namespace:         os.Getenv("REDIS_NAMESPACE"),
	}
	return config, nil
}

func PRSrvLoad() (*RedisServiceConfig, error) {
	if err := godotenv.Load(".env"); err != nil {
		logrus.Error("Failed to load .env file: ", err)
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	toDuration := func(key string) time.Duration {
		secs, _ := strconv.Atoi(os.Getenv(key))
		return time.Duration(secs) * time.Second
	}

	config := &RedisServiceConfig{
		MaxConnectionsPerUser: os.Getenv("MAX_CONNECTIONS_PER_USER"),
		RateLimitPerUser:      toDuration("REDIS_RATE_LIMIT"),
		CleanupInterval:       toDuration("REDIS_CLEANUP_INTERVAL"),
		IdleThreshold:         toDuration("REDIS_IDLE_THRESHOLD"),
	}
	return config, nil
}
