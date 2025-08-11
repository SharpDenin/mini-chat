package utils

import (
	"net/http"
	"strings"

	pb "profile_service/internal/app/auth/gRPC"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NewAuthMiddleware(authServer pb.AuthServiceServer, log *logrus.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if isPublicEndpoint(ctx) {
			ctx.Next()
			return
		}

		// Извлекаем токен
		token := extractToken(ctx)
		if token == "" {
			err := NewCustomError(http.StatusUnauthorized, "Authorization token is missing", nil)
			HandleError(ctx, err, log)
			ctx.Abort()
			return
		}

		resp, err := authServer.ValidateToken(ctx, &pb.TokenRequest{Token: token})
		if err != nil || !resp.Valid {
			customErr := NewCustomError(http.StatusUnauthorized, "Invalid token", err)
			HandleError(ctx, customErr, log)
			ctx.Abort()
			return
		}

		ctx.Set("user_id", resp.UserId)
		ctx.Next()
	}
}

func isPublicEndpoint(ctx *gin.Context) bool {
	publicEndpoints := map[string]bool{
		"/api/v1/auth/login":    true,
		"/api/v1/auth/register": true,
		"/swagger/*any":         true,
	}
	return publicEndpoints[ctx.Request.URL.Path]
}

func extractToken(ctx *gin.Context) string {
	header := ctx.GetHeader("Authorization")
	if len(header) > 7 && strings.EqualFold(header[:7], "Bearer ") {
		return header[7:]
	}
	return ""
}
