package di

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DBHandler struct {
	logger *zap.Logger
	pool   *pgxpool.Pool
}

func NewDBHandler(ctx context.Context, logger *zap.Logger, databaseURL string) (*DBHandler, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL cannot be empty")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		logger.Error("failed to create connection pool", zap.Error(err))
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection with a ping
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		logger.Error("failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection established")
	return &DBHandler{
		logger: logger,
		pool:   pool,
	}, nil
}

func (h *DBHandler) Pool() *pgxpool.Pool {
	return h.pool
}

func (h *DBHandler) Close() {
	if h.pool != nil {
		h.pool.Close()
		h.logger.Info("database connection closed")
	}
}

func (h *DBHandler) Ping(ctx context.Context) error {
	return h.pool.Ping(ctx)
}
