package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Jwt        string
	Host       string
	User       string
	Password   string
	Sslmode    string
	RoomDbName string
	RoomDbPort string
}

func Load() (*Config, error) {
	if err := godotenv.Load("../.env"); err != nil {
		logrus.Error("Failed to load .env file: ", err)
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}
	config := &Config{
		Jwt:        os.Getenv("JWT_SECRET"),
		Host:       os.Getenv("HOST"),
		User:       os.Getenv("USER"),
		Password:   os.Getenv("PASSWORD"),
		Sslmode:    os.Getenv("SSLMODE"),
		RoomDbName: os.Getenv("ROOM_DBNAME"),
		RoomDbPort: os.Getenv("ROOM_DBPORT"),
	}
	return config, nil
}
