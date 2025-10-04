package main

import (
	"context"
	"net"
	_ "profile_service/docs"
	"profile_service/internal/config"
	"profile_service/internal/repository/db"
	"profile_service/internal/repository/profile_repo"
	"profile_service/internal/service"
	"profile_service/internal/transport"
	"profile_service/middleware_profile"
	"profile_service/pkg/grpc_generated/profile"
	"profile_service/pkg/grpc_server"
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
// @host localhost:8083
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

	userRepo := profile_repo.NewProfileRepo(database.DB, log)
	userService := service.NewUserService(userRepo, log)

	authServer := grpc_server.NewAuthServer(log, userService, cfg.Jwt)
	directoryServer := grpc_server.NewDirectoryServer(log, userService)

	// ЗАПУСК gRPC СЕРВЕРОВ ПЕРВЫМИ
	log.Info("Starting gRPC servers...")

	// Auth gRPC server
	authListener, err := net.Listen("tcp", "0.0.0.0:50053")
	if err != nil {
		log.Fatalf("Failed to listen on auth port 50053: %v", err)
	}
	authGrpcServer := grpc.NewServer()
	profile.RegisterAuthServiceServer(authGrpcServer, authServer)
	go func() {
		log.Info("gRPC auth server starting on :50053")
		if err := authGrpcServer.Serve(authListener); err != nil {
			log.Fatalf("Failed to serve gRPC auth: %v", err)
		}
	}()

	// Directory gRPC server
	dirListener, err := net.Listen("tcp", "0.0.0.0:50054")
	if err != nil {
		log.Fatalf("Failed to listen on directory port 50054: %v", err)
	}
	dirGrpcServer := grpc.NewServer()
	profile.RegisterUserDirectoryServer(dirGrpcServer, directoryServer)
	go func() {
		log.Info("gRPC directory server starting on :50054")
		if err := dirGrpcServer.Serve(dirListener); err != nil {
			log.Fatalf("Failed to serve gRPC directory: %v", err)
		}
	}()

	// Даем время gRPC серверам запуститься
	log.Info("Waiting for gRPC servers to start...")
	time.Sleep(2 * time.Second)

	// ЗАТЕМ запускаем HTTP сервер
	userHandler := transport.NewUserHandler(userService, authServer, log)
	authMiddleware := middleware_profile.NewAuthMiddleware(authServer, log)

	router := gin.Default()
	router.Use(
		gin.Recovery(),
		middleware_profile.ErrorMiddleware(log),
	)

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
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Info("Starting HTTP server on :8083")
	if err := router.Run(":8083"); err != nil {
		log.Fatal("Failed to start HTTP server: ", err)
	}
}
