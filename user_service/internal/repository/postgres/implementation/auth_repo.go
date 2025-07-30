package implementation

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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