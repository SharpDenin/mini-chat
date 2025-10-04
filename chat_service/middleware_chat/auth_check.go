package middleware_chat

import (
	"chat_service/pkg/grpc_generated/profile"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ProfileClient interface {
	ValidateToken(ctx context.Context, req *profile.TokenRequest) (*profile.TokenResponse, error)
}

func NewAuthMiddleware(client ProfileClient, log *logrus.Logger) gin.HandlerFunc {
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

		// Проверяем токен через gRPC-клиент
		resp, err := client.ValidateToken(ctx, &profile.TokenRequest{Token: token})
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
	return strings.HasPrefix(ctx.Request.URL.Path, "/swagger/")
}

func extractToken(ctx *gin.Context) string {
	header := ctx.GetHeader("Authorization")
	if len(header) > 7 && strings.EqualFold(header[:7], "Bearer ") {
		return header[7:]
	}
	return ""
}
