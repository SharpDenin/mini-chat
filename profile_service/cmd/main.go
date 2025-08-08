package main

import (
	"context"
	"net"
	_ "profile_service/cmd/docs"
	pb "profile_service/internal/app/auth/gRPC"
	auth "profile_service/internal/app/auth/grpc_server"
	"profile_service/internal/app/user/delivery/http"
	"profile_service/internal/app/user/service"
	"profile_service/internal/config"
	"profile_service/internal/repository/db"
	"profile_service/internal/repository/profile_repo"
	"profile_service/internal/utils"

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
func main() {
	gin.SetMode(gin.ReleaseMode)
	log := logrus.New()
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

	userRepo := profile_repo.NewProfileRepo(database.DB, log)
	userService := service.NewUserService(userRepo, log)

	authServer := auth.NewAuthServer(log, userService, cfg.Jwt)
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	go func() {
		log.Printf("gRPC server running on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	userHandler, conn := http.NewUserHandler(userService, cfg.GRPCPort, log)
	if userHandler == nil {
		log.Fatal("Failed to create user handler")
	}
	defer conn.Close()

	router := gin.Default()
	router.Use(
		gin.Recovery(),
		utils.ErrorMiddleware(log),
	)

	api := router.Group("/api/v1")
	{
		authUser := api.Group("/auth")
		{
			authUser.POST("/login", userHandler.PostLogin)
		}
		users := api.Group("/users")
		{
			users.POST("", userHandler.PostUser)
			users.GET("", userHandler.GetFilteredUserList)
			users.GET("/:id", userHandler.GetUserById)
			users.PUT("/:id", userHandler.PutUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Запуск сервера
	log.Info("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
