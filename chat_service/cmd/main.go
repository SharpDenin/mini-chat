package main

import (
	"chat_service/internal/config"
	"chat_service/internal/repository/db"
	"chat_service/pkg/grpc_client"
	"context"
	"proto/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	log := logrus.New()
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка получения конфигурации %w", err)
	}

	database, err := db.NewDB(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	//roomRepo := room_repo.NewRoomRepo(database.DB, log)

	profileClient, err := grpc_client.NewProfileClient("profileService:50051", "profileService:50052")
	if err != nil {
		log.Fatalf("failed to create profile client: %v", err)
	}
	defer func() {
		if err := profileClient.Close(); err != nil {
			log.Printf("failed to close profile client: %v", err)
		}
	}()

	//roomService := service.NewRoomService(profileClient, roomRepo, log)

	// Создаем Gin-роутер
	router := gin.Default()

	router.Use(middleware.NewAuthMiddleware(profileClient, log))

}
