package config

import (
	"fmt"
	"os"
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
	config := &Config{
		Jwt:        os.Getenv("JWT_SECRET"),
		Host:       os.Getenv("HOST"),
		User:       os.Getenv("USER"),
		Password:   os.Getenv("PASSWORD"),
		Sslmode:    os.Getenv("SSLMODE"),
		RoomDbName: os.Getenv("ROOM_DBNAME"),
		RoomDbPort: os.Getenv("ROOM_DBPORT"),
	}
	if config.Host == "" {
		return nil, fmt.Errorf("HOST environment variable is required")
	}
	return config, nil
}
