package main

import (
	_ "chat_service/docs"
	pConfig "chat_service/internal/presence/config"
	pRepo "chat_service/internal/presence/repository"
	pService "chat_service/internal/presence/service"
	rConfig "chat_service/internal/room/config"
	rRepo "chat_service/internal/room/repository"
	"chat_service/internal/room/repository/db"
	rService "chat_service/internal/room/service"
	"chat_service/middleware_chat"
	"chat_service/pkg/grpc_client"
	"chat_service/transport"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// Инициализация логгера и контекста
	gin.SetMode(gin.ReleaseMode)
	log := logrus.New()
	ctx := context.Background()

	// Загрузка конфигурации room/roomMember-сервиса
	rCfg, err := rConfig.Load()
	if err != nil {
		log.Fatal("Ошибка получения конфигурации (room) %w", err)
	}

	// Загрузка конфигурации presence-репозитория
	prCfg, err := pConfig.PRCfgLoad()
	if err != nil {
		log.Fatal("Ошибка получения конфигурации (presence repo) %w", err)
	}

	// Загрузка конфигурации presence-сервиса
	psCfg, err := pConfig.PRSrvLoad()
	if err != nil {
		log.Fatal("Ошибка получения конфигурации (presence service) %w", err)
	}

	// Инициализация gRPC-клиента (ProfileClient)
	profileClient, err := grpc_client.NewProfileClient("localhost:50053", "localhost:50054")
	if err != nil {
		log.Fatalf("failed to create profile client: %v", err)
	}
	defer func() {
		if err := profileClient.Close(); err != nil {
			log.Printf("failed to close profile client: %v", err)
		}
	}()

	// Инициализация БД room/roomMember-сервиса
	database, err := db.NewDB(ctx, rCfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Миграция БД room/roomMember-сервиса
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Инициализация room/roomMember-репозиториев
	roomRepo := rRepo.NewRoomRepo(database.DB, log)
	roomMemberRepo := rRepo.NewRoomMemberRepo(database.DB, log)

	// Инициализация presence-репозитория
	presenceRepo, err := pRepo.NewRedisRepo(prCfg)

	// Инициализация room/roomMember-сервисов
	roomService := rService.NewRoomService(profileClient, roomRepo, roomMemberRepo, database.DB, log)
	roomMemberService := rService.NewRoomMemberService(profileClient, roomRepo, roomMemberRepo, database.DB, log)

	// Инициализация presence-сервиса
	presenceService := pService.NewPresenceService(presenceRepo, log, psCfg, nil, nil)

	// Запуск cleanup-горутины
	go func() {
		ticker := time.NewTicker(psCfg.CleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			if err := presenceService.CleanupStaleData(ctx); err != nil {
				log.Errorf("failed to cleanup stale presence: %v", err)
			}
		}
	}()

	// Инициализация room/roomMember-хэндлера
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
		roomMember := api.Group("/room-member")
		{
			roomMember.GET("/rooms/:room_id/members", roomHandler.GetMemberList)
			roomMember.POST("/rooms/:room_id/members/:user_id", roomHandler.CreateRoomMember)
			roomMember.PUT("/rooms/:room_id/members/:user_id/admin", roomHandler.SetAdminMember)
			roomMember.DELETE("/rooms/:room_id/members/:user_id", roomHandler.DeleteRoomMember)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	srv := &http.Server{
		Addr:         ":8084",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	serverErr := make(chan error, 1)

	go func() {
		log.Info("Starting server on :8084")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Info("Received shutdown signal...")
	case err := <-serverErr:
		log.Errorf("Server error: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Info("Shutting down server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exiting properly")
}
