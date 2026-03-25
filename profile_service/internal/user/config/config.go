package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Jwt           string
	Host          string
	User          string
	Password      string
	Sslmode       string
	ProfileDbname string
	ProfilePort   string
	GRPCPort      string
}

func Load() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		logrus.Error("Failed to load .env file: ", err)
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}
	config := &Config{
		Jwt: os.Getenv("JWT_SECRET"),
		//Host:          os.Getenv("HOST"),
		Host:          os.Getenv("HOST_PROD"),
		User:          os.Getenv("USER"),
		Password:      os.Getenv("PASSWORD"),
		Sslmode:       os.Getenv("SSLMODE"),
		ProfileDbname: os.Getenv("PROFILE_DBNAME"),
		// ProfilePort:   os.Getenv("PROFILE_PORT"),
		ProfilePort: os.Getenv("PROFILE_PORT_PROD"),
		GRPCPort:    os.Getenv("GRPC_PORT"),
	}
	return config, nil
}
