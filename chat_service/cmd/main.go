package main

import (
	_ "chat_service/docs"
	transport "chat_service/http"
	"chat_service/internal/authz"
	pConfig "chat_service/internal/presence/config"
	pRepo "chat_service/internal/presence/repository"
	"chat_service/internal/presence/service"
	"chat_service/internal/pubsub"
	rConfig "chat_service/internal/room/config"
	rRepo "chat_service/internal/room/repository"
	"chat_service/internal/room/repository/db"
	rService "chat_service/internal/room/service"
	"chat_service/internal/websocket"
	"chat_service/internal/websocket/dto"
	"chat_service/internal/websocket/handler"
	"chat_service/middleware_chat"
	"chat_service/pkg/grpc_client"
	"chat_service/pkg/grpc_generated/chat"
	"chat_service/pkg/grpc_server"
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
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
	ctx, cancelMain := context.WithCancel(context.Background())
	defer cancelMain()

	// Загрузка конфигурации room/roomMember-сервиса
	rCfg, err := rConfig.Load()
	if err != nil {
		log.Fatal("Ошибка получения конфигурации (room) %w", err)
	}

	// Инициализация БД
	database, err := db.NewDB(ctx, rCfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Миграция БД
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
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

	// Инициализация репозиториев и сервисов
	roomRepo := rRepo.NewRoomRepo(database.DB, log)
	roomMemberRepo := rRepo.NewRoomMemberRepo(database.DB, log)

	roomService := rService.NewRoomService(profileClient, roomRepo, roomMemberRepo, database.DB, log)
	roomMemberService := rService.NewRoomMemberService(profileClient, roomRepo, roomMemberRepo, database.DB, log)
	authzService := authz.NewGrpcAuthz(profileClient)

	// Загрузка конфигурации redis-модуля
	redisCfg, err := pConfig.RedisCfgLoad()
	if err != nil {
		log.Fatal("Ошибка получения конфигурации (presence service) %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.RedisDb,
	})
	defer rdb.Close()

	// Инициализация presence-репозитория
	presenceRepo := pRepo.NewPresenceRepo(rdb, redisCfg.IdleThreshold)

	// Инициализация сервисов с использованием redis
	bus := service.NewPresenceEventBus()
	presenceService := service.NewPresenceService(presenceRepo, bus, redisCfg)
	pb := pubsub.NewRedisPubSub(rdb)

	// Инициализация gRPC-сервера
	presenceServer := grpc_server.NewGRPCServer(presenceService)

	// Запуск gRPC-сервера
	log.Info("Starting gRPC server...")

	presenceListener, err := net.Listen("tcp", "0.0.0.0:50056")
	if err != nil {
		log.Fatalf("failed to listen on presence port 50056: %v", err)
	}
	presenceGrpcServer := grpc.NewServer()
	chat.RegisterPresenceServer(presenceGrpcServer, presenceServer)
	go func() {
		log.Info("gRPC presence service starting on :50056")
		if err := presenceGrpcServer.Serve(presenceListener); err != nil {
			log.Fatalf("Failed to serve gRPC auth: %v", err)
		}
	}()

	// Ожидание запуска gRPC-серверов
	log.Info("Waiting for gRPC servers to start...")
	time.Sleep(5 * time.Second)

	// Инициализация хэндлера
	roomHandler := transport.NewRoomHandler(log, roomService, roomMemberService)

	// Создание gin-роутера
	router := gin.Default()
	router.Use(
		gin.Recovery(),
		middleware_chat.ErrorMiddleware(log),
	)

	// Подписка Hub к Presence
	instance, _ := os.Hostname()
	hub := websocket.NewHub(bus.Subscribe(), pb, instance)
	go hub.Run(ctx)

	// Инициализация ws-роутера, регистрация хэндлеров и апгрейд соединения
	wsRouter := websocket.NewRouter()
	wsRouter.Register(dto.MessagePresence, handler.PresenceHandler)
	wsRouter.Register(dto.MessageChat, handler.ChatHandler)
	wsHandler := handler.NewWSHandler(ctx, wsRouter, hub, presenceService, authzService, profileClient)
	router.GET("/ws", gin.WrapF(wsHandler))

	// Регистрация методов API
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

	// Объявление параметров запуска сервера + graceful shutdown
	srv := &http.Server{
		Addr:    ":8084",
		Handler: router,
		// Настройки timeouts для защиты от медленных клиентов
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	serverErr := make(chan error, 1)

	// Запуск сервера в горутине
	go func() {
		log.Info("Starting server on :8084")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Отправка сигнала об остановке сервера в канал
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Info("Received shutdown signal...")
	case err := <-serverErr:
		log.Errorf("Server error: %v", err)
	}

	// Задание контекста с ожиданием timeout через 30сек (завершение всех запросов, находящихся в обработке)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Остановка сервера
	log.Info("Shutting down server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	cancelMain()
	presenceGrpcServer.GracefulStop()
	_ = rdb.Close()

	log.Info("Server exiting properly")
}
