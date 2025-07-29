package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	Host       string
	User       string
	Password   string
	Sslmode    string
	UserDbname string
	UserPort   string
}

func Load() (*Config, error) {
	if err := godotenv.Load("../.env"); err != nil {
		logrus.Error("Failed to load .env file: ", err)
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}
	config := &Config{
		Host:       os.Getenv("HOST"),
		User:       os.Getenv("USER"),
		Password:   os.Getenv("PASSWORD"),
		Sslmode:    os.Getenv("SSLMODE"),
		UserDbname: os.Getenv("USER_DBNAME"),
		UserPort:   os.Getenv("USER_PORT"),
	}
	return config, nil
}
