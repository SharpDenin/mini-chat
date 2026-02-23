package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type RedisConfig struct {
	Addr          string
	Password      string
	RedisDb       int
	IdleThreshold time.Duration
}

func RedisCfgLoad() (*RedisConfig, error) {
	if err := godotenv.Load(".env"); err != nil {
		logrus.Error("Failed to load .env file: ", err)
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	toDuration := func(key string) time.Duration {
		secs, _ := strconv.Atoi(os.Getenv(key))
		return time.Duration(secs) * time.Second
	}

	toInt := func(key string) int {
		num, _ := strconv.Atoi(os.Getenv(key))
		return num
	}

	config := &RedisConfig{
		Addr:          os.Getenv("REDIS_ADDR"),
		Password:      os.Getenv("REDIS_PASSWORD"),
		RedisDb:       toInt("REDIS_DB"),
		IdleThreshold: toDuration("IDLE_THRESHOLD"),
	}
	return config, nil
}
