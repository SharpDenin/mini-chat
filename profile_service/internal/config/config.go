package config

import (
	"os"
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
	config := &Config{
		Jwt:           os.Getenv("JWT_SECRET"),
		Host:          os.Getenv("HOST"),
		User:          os.Getenv("USER"),
		Password:      os.Getenv("PASSWORD"),
		Sslmode:       os.Getenv("SSLMODE"),
		ProfileDbname: os.Getenv("PROFILE_DBNAME"),
		ProfilePort:   os.Getenv("PROFILE_PORT"),
		GRPCPort:      os.Getenv("GRPC_PORT"),
	}

	return config, nil
}
