package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"time"
)

type KafkaConfig struct {
	Brokers      []string
	Token        string
	Topic        string
	ClientId     string
	BatchSize    int
	BatchTimeout time.Duration
	RetryMax     int
	RetryBackoff time.Duration
	RequiredAcks int
	Compression  string
}

func KafkaCfgLoad() (*KafkaConfig, error) {
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

	brokersStr := os.Getenv("BROKER")
	var brokers []string
	if brokersStr != "" {
		brokers = strings.Split(brokersStr, ",")
	}

	config := &KafkaConfig{
		Brokers:      brokers,
		Token:        os.Getenv("TOKEN"),
		Topic:        os.Getenv("TOPIC"),
		ClientId:     os.Getenv("CLIENT_ID"),
		BatchSize:    toInt(os.Getenv("BATCH_SIZE")),
		BatchTimeout: toDuration(os.Getenv("BATCH_TIMEOUT")),
		RetryMax:     toInt(os.Getenv("RETRY_MAX")),
		RetryBackoff: toDuration(os.Getenv("RETRY_BACKOFF")),
		RequiredAcks: toInt(os.Getenv("REQUIRED_ACKS")),
		Compression:  os.Getenv("COMPRESSION"),
	}

	return config, nil
}
