package main

import (
	_ "chat_service/docs"
	"chat_service/internal/room/room_config"
	"chat_service/internal/room/room_repository"
	"chat_service/internal/room/room_repository/db"
	"chat_service/internal/room/room_service"
	"chat_service/internal/transport"
	"chat_service/middleware_chat"
	"chat_service/pkg/grpc_client"
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
	gin.SetMode(gin.ReleaseMode)
	log := logrus.New()
	ctx := context.Background()
	cfg, err := room_config.Load()
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

	roomRepo := room_repository.NewRoomRepo(database.DB, log)
	roomMemberRepo := room_repository.NewRoomMemberRepo(database.DB, log)

	profileClient, err := grpc_client.NewProfileClient("localhost:50053", "localhost:50054")
	if err != nil {
		log.Fatalf("failed to create profile client: %v", err)
	}
	defer func() {
		if err := profileClient.Close(); err != nil {
			log.Printf("failed to close profile client: %v", err)
		}
	}()

	roomService := room_service.NewRoomService(profileClient, roomRepo, roomMemberRepo, database.DB, log)
	roomMemberService := room_service.NewRoomMemberService(profileClient, roomRepo, roomMemberRepo, database.DB, log)

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
