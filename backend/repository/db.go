package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/agentrading/backend/config"
)

var Pool *pgxpool.Pool

// Connect initializes the PostgreSQL connection pool.
func Connect() error {
	cfg, err := pgxpool.ParseConfig(config.AppConfig.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	Pool = pool
	log.Println("[DB] Connected to PostgreSQL successfully")
	return nil
}

// Close gracefully shuts down the pool.
func Close() {
	if Pool != nil {
		Pool.Close()
		log.Println("[DB] PostgreSQL connection pool closed")
	}
}
