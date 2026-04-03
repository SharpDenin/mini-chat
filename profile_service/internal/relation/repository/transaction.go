package repository

import (
	"context"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TransactionManagerInterface interface {
	RunInTransaction(ctx context.Context, fn func(tx FriendshipRepositoryInterface) error) error
}

type TransactionManager struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewTransactionManager(db *gorm.DB, log *logrus.Logger) TransactionManagerInterface {
	return &TransactionManager{
		db:  db,
		log: log,
	}
}

func (t *TransactionManager) RunInTransaction(ctx context.Context, fn func(tx FriendshipRepositoryInterface) error) error {
	logger := t.log.WithContext(ctx)

	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := NewFriendshipRepository(tx, logger)
		return fn(repo)
	})
}
