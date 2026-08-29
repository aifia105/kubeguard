package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitPool(ctx context.Context) error {
	connString := os.Getenv("KUBEGUARD_POSTGRES_URL")
	if connString == "" {
		return fmt.Errorf("KUBEGUARD_POSTGRES_URL environment variable is not set")
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to reach database: %w", err)
	}
	Pool = pool
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
