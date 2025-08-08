package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Host          string
	User          string
	Password      string
	Sslmode       string
	ProfileDbname string
	ProfilePort   string
}

func Load() (*Config, error) {
	if err := godotenv.Load("../.env"); err != nil {
		logrus.Error("Failed to load .env file: ", err)
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}
	config := &Config{
		Host:          os.Getenv("HOST"),
		User:          os.Getenv("USER"),
		Password:      os.Getenv("PASSWORD"),
		Sslmode:       os.Getenv("SSLMODE"),
		ProfileDbname: os.Getenv("PROFILE_DBNAME"),
		ProfilePort:   os.Getenv("PROFILE_PORT"),
	}
	return config, nil
}
