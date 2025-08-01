package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
	"user_service/internal/config"
)

type Database struct {
	DB      *gorm.DB
	Migrate *migrate.Migrate
}

func NewDB(ctx context.Context, cfg *config.Config) (*Database, error) {
	dsn :=
		`host=` + cfg.Host +
			` user=` + cfg.User +
			` password=` + cfg.Password +
			` dbname=` + cfg.UserDbname +
			` port=` + cfg.UserPort +
			` sslmode=` + cfg.Sslmode

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connection failed %w", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	migrationPath := filepath.Join("migration")
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory not found: %w", err)
	}

	m, err := migrate.New(
		"file://"+migrationPath,
		"postgres://postgres:postgres@localhost:5460/user_service_db?sslmode=disable",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrate: %w", err)
	}

	return &Database{
		DB:      gormDB,
		Migrate: m,
	}, nil
}

func (d *Database) Close() error {
	if _, err := d.Migrate.Close(); err != nil {
		return fmt.Errorf("failed to close migrate: %w", err)
	}

	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	return sqlDB.Close()
}

func (d *Database) RunMigrations() error {
	log.Println("Applying database migrations...")
	err := d.Migrate.Up()

	switch {
	case err == nil:
		log.Println("All migrations applied successfully")
		return nil
	case errors.Is(err, migrate.ErrNoChange):
		log.Println("No new migrations to apply")
		return nil
	default:
		return fmt.Errorf("migration failed: %w", err)
	}
}
