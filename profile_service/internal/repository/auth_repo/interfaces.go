package auth_repo

import (
	"context"
	"profile_service/internal/app/auth/model"
)

type AuthRepoInterface interface {
	SaveToken(ctx context.Context, token *model.AuthToken) error
	GetToken(ctx context.Context, tokenString string) (*model.AuthToken, error)
	RevokeToken(ctx context.Context, tokenString string) error
	RevokeAllUserTokens(ctx context.Context, userId int64) error
	DeleteExpiredTokens(ctx context.Context) error
}
