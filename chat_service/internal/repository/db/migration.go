package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

func RunMigration(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		log.Println("Context error:", err)
	}

	m, err := migrate.New("file://migration", "postgres://postgres:postgres@chat-db:5432/chat_service_db?sslmode=disable")
	if err != nil {
		log.Println("Migration error:", err)
		return fmt.Errorf("error migrating database: %w", err)
	}
	defer func() {
		if _, err := m.Close(); err != nil {
			log.Println("Migration close error:", err)
		}
	}()
	fmt.Println("Applying database migrations...")

	err = m.Up()

	switch {
	case err == nil:
		fmt.Println("All migrations applied successfully")
		return nil
	case errors.Is(err, migrate.ErrNoChange):
		fmt.Println("No new migrations to apply")
		return nil
	default:
		return fmt.Errorf("migration up error: %w", err)
	}
}
