package config

import (
	"os"
	"strconv"
	"time"
)

type RedisConfig struct {
	Addr          string
	Password      string
	RedisDb       int
	IdleThreshold time.Duration
}

func RedisCfgLoad() (*RedisConfig, error) {
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
