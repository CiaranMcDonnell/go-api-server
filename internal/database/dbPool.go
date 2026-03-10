package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DBPool *pgxpool.Pool

func InitializeConnectionPool(ctx context.Context, config *utils.Config, runMigrations bool) (*pgxpool.Pool, error) {
	if runMigrations {
		db, err := RunMigrations()
		if err != nil {
			return nil, fmt.Errorf("could not run migrations: %w", err)
		}
		defer db.Close()
		slog.Info("Database migrations applied successfully")
	}

	poolConfig, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %w", err)
	}

	if config.DBMaxConns > 0 {
		poolConfig.MaxConns = int32(config.DBMaxConns)
	}
	if config.DBMinConns > 0 {
		poolConfig.MinConns = int32(config.DBMinConns)
	}
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.MaxConnLifetime = 1 * time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DBPool = pool
	slog.Info("PostgreSQL connection pool established",
		"max_conns", poolConfig.MaxConns,
		"min_conns", poolConfig.MinConns,
	)

	return pool, nil
}

func CloseDB() {
	if DBPool != nil {
		DBPool.Close()
		slog.Info("PostgreSQL connection pool closed")
	}
}

func HealthCheck(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := DBPool.Ping(timeoutCtx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}
