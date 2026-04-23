package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	_ "profile_service/docs"
	transport "profile_service/http"
	"profile_service/internal/config"
	"profile_service/internal/config/db"
	"profile_service/internal/kafka/friendship_producer"
	relRepo "profile_service/internal/relation/repository"
	relService "profile_service/internal/relation/service"
	"profile_service/internal/user/cache"
	userRepo "profile_service/internal/user/repository"
	userService "profile_service/internal/user/service"
	"profile_service/middleware_profile"
	"profile_service/pkg/grpc_client"
	"profile_service/pkg/grpc_generated/profile"
	"profile_service/pkg/grpc_server"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title ProfileService API
// @version 1.0
// @description API для управления пользователями
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
func main() {

	// Инициализация логгера и контекста
	gin.SetMode(gin.ReleaseMode)
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)

	ctx, cancelMain := context.WithCancel(context.Background())
	defer cancelMain()

	// Загрузка конфигурации user-сервиса
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка получения конфигурации %w", err)
	}

	// Загрузка конфигурации kafka
	kafkaCfg, err := config.KafkaCfgLoad()
	if err != nil {
		log.Fatal("Ошибка получения конфигурации Kafka: ", err)
	}

	// Инициализация БД user-сервиса
	database, err := db.NewDB(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Миграция БД user-сервиса
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Инициализация gRPC-клиента (PresenceClient)
	presenceAddr := os.Getenv("CHAT_PRESENCE_GRPC_ADDR")
	if presenceAddr == "" {
		presenceAddr = "chat-service:50056"
	}
	presenceClient, err := grpc_client.NewPresenceClient(presenceAddr)
	if err != nil {
		log.Fatalf("failed to create profile client: %v", err)
	}
	defer func() {
		if err = presenceClient.Close(); err != nil {
			log.Printf("failed to close presence client: %v", err)
		}
	}()

	// Инициализация кэширования статусов
	presenceCache := cache.NewPresenceCache(3 * time.Second)
	defer presenceCache.Stop()

	// Инициализация user-репозитория и сервисов
	userRepo := userRepo.NewProfileRepo(database.DB, log)
	userService := userService.NewUserService(userRepo, presenceClient, presenceCache, log)

	// Инициализация репозиториев для дружбы
	friendshipRepo := relRepo.NewFriendshipRepository(database.DB, log.WithField("component", "friendship_repo"))
	txManager := relRepo.NewTransactionManager(database.DB, log)

	// Инициализация Kafka friendship_producer
	kafkaProducer, err := friendship_producer.NewKafkaProducer(kafkaCfg.Brokers, kafkaCfg.Topic)
	if err != nil {
		log.Fatalf("Failed to run kafkaProducer: %v", err)
	}
	defer kafkaProducer.Close()

	// Инициализация Outbox friendship_producer
	outboxProducer := friendship_producer.NewOutboxProducer(database.DB, kafkaProducer, 100)
	defer outboxProducer.Close()

	// Инициализация сервисов дружбы
	relationChecker := relService.NewRelationChecker(userService, friendshipRepo)
	friendshipService := relService.NewFriendshipService(
		friendshipRepo,
		userService,
		txManager,
		outboxProducer,
		log.WithField("component", "friendship_service"),
	)

	// Инициализация gRPC-серверов
	authServer := grpc_server.NewAuthServer(log, userService, cfg.Jwt)
	directoryServer := grpc_server.NewDirectoryServer(log, userService)
	authzServer := grpc_server.NewAuthorizationServer(relationChecker)

	// Запуск gRPC серверов
	log.Info("Starting gRPC servers...")

	// Auth gRPC-server
	authPort := os.Getenv("PROFILE_GRPC_AUTH_PORT")
	if authPort == "" {
		authPort = "50053"
	}
	authAddr := "0.0.0.0:" + authPort
	authListener, err := net.Listen("tcp", authAddr)
	if err != nil {
		log.Fatalf("Failed to listen on auth port %s: %v", authPort, err)
	}
	authGrpcServer := grpc.NewServer()
	profile.RegisterAuthServiceServer(authGrpcServer, authServer)
	profile.RegisterAuthorizationServiceServer(authGrpcServer, authzServer)
	go func() {
		log.Infof("gRPC auth server starting on :%s", authPort)
		if err := authGrpcServer.Serve(authListener); err != nil {
			log.Fatalf("Failed to serve gRPC auth: %v", err)
		}
	}()

	// Directory gRPC-server
	dirPort := os.Getenv("PROFILE_GRPC_DIRECTORY_PORT")
	if dirPort == "" {
		dirPort = "50054"
	}
	dirAddr := "0.0.0.0:" + dirPort
	dirListener, err := net.Listen("tcp", dirAddr)
	if err != nil {
		log.Fatalf("Failed to listen on directory port %s: %v", dirPort, err)
	}
	dirGrpcServer := grpc.NewServer()
	profile.RegisterUserDirectoryServer(dirGrpcServer, directoryServer)
	go func() {
		log.Infof("gRPC directory server starting on :%s", dirPort)
		if err := dirGrpcServer.Serve(dirListener); err != nil {
			log.Fatalf("Failed to serve gRPC directory: %v", err)
		}
	}()

	// Ожидание запуска gRPC-серверов
	log.Info("Waiting for gRPC servers to start...")
	time.Sleep(5 * time.Second)

	// Инициализация хэндлера
	userHandler := transport.NewUserHandler(userService, authServer, log)
	friendshipHandler := transport.NewFriendshipHandler(
		userService,
		friendshipService,
		relationChecker,
		log,
	)

	// Подключение auth-middleware
	authMiddleware := middleware_profile.NewAuthMiddleware(authServer, log)

	// Создание gin-роутера
	router := gin.Default()
	router.Use(
		gin.Recovery(),
		middleware_profile.ErrorMiddleware(log),
		middleware_profile.NewCORS(middleware_profile.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}),
	)

	// Регистрация методов API
	api := router.Group("/api/v1")
	{
		authUser := api.Group("/auth")
		{
			authUser.POST("/login", userHandler.PostLogin)
			authUser.POST("/register", userHandler.PostUser)
		}
		users := api.Group("/users")
		users.Use(authMiddleware)
		{
			users.GET("", userHandler.GetFilteredUserList)
			users.GET("/:id", userHandler.GetUserById)
			users.PUT("/:id", userHandler.PutUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
		friendRequests := api.Group("/friends/requests")
		friendRequests.Use(authMiddleware)
		{
			friendRequests.POST(":receiver_id", friendshipHandler.PostFriendRequest)
			friendRequests.PUT("/:request_id", friendshipHandler.AnswerFriendRequest)
			friendRequests.DELETE("/:request_id", friendshipHandler.CancelFriendRequest)
			friendRequests.GET("/state", friendshipHandler.CheckRequestState)
		}
		friends := api.Group("/friends")
		friends.Use(authMiddleware)
		{
			friends.GET("", friendshipHandler.GetFriendList)
			friends.DELETE("/:friend_id", friendshipHandler.DeleteFriend)
			friends.GET("/check", friendshipHandler.CheckAreFriends)
		}
		block := api.Group("/block")
		block.Use(authMiddleware)
		{
			block.POST(":blocked_id", friendshipHandler.BlockUser)
			block.DELETE(":blocked_id", friendshipHandler.UnblockUser)
			block.GET("/:blocked_id", friendshipHandler.GetBlockInfo)
		}
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Объявление параметров запуска сервера + graceful shutdown
	srv := &http.Server{
		Addr:    ":8083",
		Handler: router,
		// Настройки timeouts для защиты от медленных клиентов
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	serverErr := make(chan error, 1)

	// Запуск сервера в горутине
	go func() {
		log.Info("Starting server on :8083")
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
	dirGrpcServer.GracefulStop()
	authGrpcServer.GracefulStop()

	log.Info("Server exiting properly")
}
