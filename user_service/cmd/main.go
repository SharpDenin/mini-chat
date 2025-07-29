package main

import (
	"context"
	"log"
	"user_service/internal/config"
	"user_service/internal/repository/db"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка получения конфигурации %w", err)
	}

	// Инициализация базы данных
	database, err := db.NewDB(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Применение миграций
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}
