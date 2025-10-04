package main

import (
	_ "chat_service/docs"
	"chat_service/internal/config"
	"chat_service/internal/repository/db"
	"chat_service/internal/repository/room_repo"
	"chat_service/internal/service"
	"chat_service/internal/transport"
	"chat_service/middleware_chat"
	"chat_service/pkg/grpc_client"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title ChatService API
// @version 1.0
// @description API для управления пользователями
// @host localhost:8084
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
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

	roomRepo := room_repo.NewRoomRepo(database.DB, log)
	roomMemberRepo := room_repo.NewRoomMemberRepo(database.DB, log)

	profileClient, err := grpc_client.NewProfileClient("localhost:50053", "localhost:50054")
	if err != nil {
		log.Fatalf("failed to create profile client: %v", err)
	}
	defer func() {
		if err := profileClient.Close(); err != nil {
			log.Printf("failed to close profile client: %v", err)
		}
	}()

	roomService := service.NewRoomService(profileClient, roomRepo, roomMemberRepo, database.DB, log)
	roomMemberService := service.NewRoomMemberService(profileClient, roomRepo, roomMemberRepo, database.DB, log)

	roomHandler := transport.NewRoomHandler(log, roomService, roomMemberService)

	// Создаем Gin-роутер
	router := gin.Default()
	router.Use(
		gin.Recovery(),
		middleware_chat.ErrorMiddleware(log),
	)

	api := router.Group("/api/v1")
	{
		api.Use(middleware_chat.NewAuthMiddleware(profileClient, log))
		room := api.Group("/room")
		{
			room.POST("", roomHandler.CreateRoom)
			room.GET("/:id", roomHandler.GetRoom)
			room.GET("", roomHandler.GetRoomList)
			room.PUT("/:id", roomHandler.RenameRoom)
			room.DELETE("/:id", roomHandler.DeleteRoom)
		}
		roomMember := api.Group("/room_member")
		{
			roomMember.POST("", roomHandler.CreateRoomMember)
			roomMember.GET("/:room_id", roomHandler.GetMemberList)
			roomMember.PUT("/:user_id/:room_id", roomHandler.SetAdminMember)
			roomMember.DELETE("/:user_id/:room_id", roomHandler.DeleteRoomMember)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Запуск сервера
	log.Info("Starting server on :8084")
	if err := router.Run(":8084"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
