package grpc_server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"profile_service/internal/service"
	"profile_service/internal/service/service_dto"
	"profile_service/middleware"
	"profile_service/pkg/grpc_generated/profile"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	log *logrus.Logger
	profile.UnimplementedAuthServiceServer
	uService  service.UserServiceInterface
	jwtSecret string
}

func NewAuthServer(log *logrus.Logger, uService service.UserServiceInterface, jwtSecret string) *AuthServer {
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

func (s *AuthServer) Register(ctx context.Context, req *profile.RegisterRequest) (*profile.RegisterResponse, error) {
	s.log.WithFields(logrus.Fields{
		"username": req.Username,
		"email":    req.Email,
	}).Debug("Register request")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.WithError(err).Error("Failed to hash password")
		return nil, status.Error(codes.Internal, "Failed to hash password")
	}

	createReq := &service_dto.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	id, err := s.uService.CreateUser(ctx, createReq)
	if err != nil {
		s.log.WithError(err).Error("Failed to create user")
		var customErr *middleware.CustomError
		if errors.As(err, &customErr) {
			return nil, status.Error(codes.Code(customErr.StatusCode), customErr.Message)
		}
		return nil, status.Error(codes.Internal, "Failed to create user")
	}

	return &profile.RegisterResponse{
		UserId:  strconv.FormatInt(id, 10),
		Message: "User registered successfully",
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *profile.LoginRequest) (*profile.LoginResponse, error) {
	s.log.WithFields(logrus.Fields{
		"username": req.Username,
	}).Debug("Login request")

	filter := service_dto.SearchUserFilter{
		Username: req.Username,
		Limit:    1,
		Offset:   0,
	}
	userList, err := s.uService.GetAllUsers(ctx, filter)
	if err != nil || len(userList.UserList) == 0 {
		s.log.WithError(err).Error("User not found")
		return nil, status.Error(codes.NotFound, "User not found")
	}
	user := userList.UserList[0]
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.log.Warn("Invalid password")
		return nil, status.Error(codes.Unauthenticated, "Invalid password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": strconv.FormatInt(user.Id, 10),
		"exp":     time.Now().Add(time.Hour * 1).Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.log.WithError(err).Error("Failed to generate token")
		return nil, status.Error(codes.Internal, "Failed to generate token")
	}
	return &profile.LoginResponse{
		Token:  tokenString,
		UserId: strconv.FormatInt(user.Id, 10),
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *profile.TokenRequest) (*profile.TokenResponse, error) {
	s.log.WithFields(logrus.Fields{
		"token": req.Token[:10] + "...",
	}).Debug("ValidateToken request")

	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		s.log.WithError(err).Warn("Invalid token")
		return &profile.TokenResponse{
			Valid: false,
			Error: "Invalid token",
		}, nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		s.log.Warn("Invalid claims")
		return &profile.TokenResponse{
			Valid: false,
			Error: "Invalid claims",
		}, nil
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		s.log.Warn("Invalid user_id in claims")
		return &profile.TokenResponse{
			Valid: false,
			Error: "Invalid user_id",
		}, nil
	}

	return &profile.TokenResponse{
		Valid:  true,
		UserId: userID,
	}, nil
}
