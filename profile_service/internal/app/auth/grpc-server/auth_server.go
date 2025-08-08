package grpc_server

import (
	"context"
	"os"
	pb "profile_service/internal/app/auth/gRPC"
	"profile_service/internal/app/user/service"

	"github.com/sirupsen/logrus"
)

type AuthServer struct {
	log *logrus.Logger
	pb.UnimplementedAuthServiceServer
	uService  service.UserService
	jwtSecret string
}

func NewAuthServer(log *logrus.Logger, uService service.UserService, jwtSecret string) *AuthServer {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &AuthServer{
		log:       log,
		uService:  uService,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return &pb.RegisterResponse{}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return &pb.LoginResponse{}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	return &pb.TokenResponse{}, nil
}
