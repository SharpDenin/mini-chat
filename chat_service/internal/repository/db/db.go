package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

func NewDB(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@chat-db:5432/chat_service_db?sslmode=disable")
	if err != nil {
		log.Println("Connection failed", err)
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		log.Println("Ping failed", err)
		pool.Close()
		return nil, err
	}
	log.Println("Successfully connected to postgres")
	return pool, nil
}
