package implementation

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
	"user_service/internal/app/auth/model"
	"user_service/internal/repository/postgres"
)

type AuthRepo struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewAuthRepo(db *gorm.DB, log *logrus.Logger) postgres.AuthRepoInterface {
	return &AuthRepo{
		db:  db,
		log: log,
	}
}

func (a *AuthRepo) SaveToken(ctx context.Context, token *model.AuthToken) error {
	if err := a.db.Create(token).Error; err != nil {
		a.log.WithError(err).Error("save token error:")
		return fmt.Errorf("save token error: %w", err)
	}
	return nil
}

func (a AuthRepo) GetToken(ctx context.Context, tokenString string) (*model.AuthToken, error) {
	var token model.AuthToken

	if err := a.db.WithContext(ctx).
		Where("token = ?", tokenString).
		First(&token).Error; err != nil {
		a.log.WithError(err).Error("get token error:")
		return nil, fmt.Errorf("get token error: %w", err)
	}
	return &token, nil
}

func (a AuthRepo) RevokeToken(ctx context.Context, tokenString string) error {
	var token model.AuthToken

	tx := a.db.Begin().WithContext(ctx)
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()
	if err := tx.WithContext(ctx).
		Where("token = ?", tokenString).
		First(&token).Error; err != nil {
		a.log.WithError(err).Error("get token error:")
		return fmt.Errorf("get token error: %w", err)
	}
	if err := tx.Model(&token).Update("revoked", true).Error; err != nil {
		a.log.WithError(err).Error("revoke token error:")
		return fmt.Errorf("revoke token error: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		a.log.WithError(err).Error("commit auth token error:")
		return fmt.Errorf("commit auth token error: %w", err)
	}
	committed = true
	return nil
}

func (a *AuthRepo) RevokeAllUserTokens(ctx context.Context, userId int64) error {
	tx := a.db.Begin().WithContext(ctx)
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()
	if err := tx.WithContext(ctx).
		Model(&model.AuthToken{}).
		Where("user_id = ? AND revoked = ?", userId, false).
		Update("revoked", true).Error; err != nil {
		a.log.WithError(err).Error("revoke all user token error:")
		return fmt.Errorf("revoke all user token error: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		a.log.WithError(err).Error("commit auth token error:")
		return fmt.Errorf("commit auth token error: %w", err)
	}
	committed = true
	return nil
}

func (a *AuthRepo) DeleteExpiredTokens(ctx context.Context) error {
	if err := a.db.Begin().WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&model.AuthToken{}).Error; err != nil {
		a.log.WithError(err).Error("delete expired token error:")
		return fmt.Errorf("delete expired token error: %w", err)
	}
	return nil
}
