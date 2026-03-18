package di

import (
	"context"
	"fmt"
	"libs/middleware"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ClickHouseClient struct {
	conn driver.Conn
}

func NewClickHouseClient(config *Config, logger *zap.Logger) (*ClickHouseClient, error) {
	opts, err := clickhouse.ParseDSN(config.ClickHouseConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ClickHouse connection string: %w", err)
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Test connection
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	logger.Info("connected to ClickHouse")

	return &ClickHouseClient{
		conn: conn,
	}, nil
}

func (c *ClickHouseClient) Close() error {
	return c.conn.Close()
}

// InsertListenEvent inserts a music listen event
func (c *ClickHouseClient) InsertListenEvent(ctx context.Context, req ListenEventRequest) error {
	logger := middleware.GetLogger(ctx)

	query := fmt.Sprintf(`
		INSERT INTO %s (
			user_uuid, music_uuid, artist_uuid, album_uuid,
			listen_duration_seconds, track_duration_seconds, completion_ratio
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, tableListenEvents)

	var albumUUID uuid.UUID
	if req.AlbumUUID != nil {
		albumUUID = *req.AlbumUUID
	}

	err := c.conn.Exec(ctx, query,
		req.UserUUID,
		req.MusicUUID,
		req.ArtistUUID,
		albumUUID,
		req.ListenDurationSeconds,
		req.TrackDurationSeconds,
		req.CompletionRatio,
	)
	if err != nil {
		logger.Error("failed to insert listen event",
			zap.Error(err),
			zap.String("music_uuid", req.MusicUUID.String()),
		)
		return fmt.Errorf("failed to insert listen event: %w", err)
	}

	logger.Debug("inserted listen event",
		zap.String("music_uuid", req.MusicUUID.String()),
	)
	return nil
}

// InsertLikeEvent inserts a music like event
func (c *ClickHouseClient) InsertLikeEvent(ctx context.Context, req LikeEventRequest) error {
	logger := middleware.GetLogger(ctx)

	query := fmt.Sprintf(`
		INSERT INTO %s (user_uuid, music_uuid, artist_uuid) VALUES (?, ?, ?)
	`, tableLikeEvents)

	err := c.conn.Exec(ctx, query,
		req.UserUUID,
		req.MusicUUID,
		req.ArtistUUID,
	)
	if err != nil {
		logger.Error("failed to insert like event",
			zap.Error(err),
			zap.String("music_uuid", req.MusicUUID.String()),
		)
		return fmt.Errorf("failed to insert like event: %w", err)
	}

	logger.Debug("inserted like event",
		zap.String("music_uuid", req.MusicUUID.String()),
	)

	return nil
}

// UpsertTheme inserts/updates a music theme (ReplacingMergeTree)
func (c *ClickHouseClient) UpsertTheme(ctx context.Context, req ThemeEventRequest) error {
	logger := middleware.GetLogger(ctx)

	query := fmt.Sprintf(`
		INSERT INTO %s (music_uuid, theme) VALUES (?, ?)
	`, tableTheme)

	err := c.conn.Exec(ctx, query, req.MusicUUID, req.Theme)
	if err != nil {
		logger.Error("failed to upsert theme",
			zap.Error(err),
			zap.String("music_uuid", req.MusicUUID.String()),
			zap.String("theme", req.Theme),
		)
		return fmt.Errorf("failed to upsert theme: %w", err)
	}

	logger.Debug("upserted theme",
		zap.String("music_uuid", req.MusicUUID.String()),
		zap.String("theme", req.Theme),
	)

	return nil
}

// UpsertUserDim inserts/updates a user dimension (ReplacingMergeTree)
func (c *ClickHouseClient) UpsertUserDim(ctx context.Context, req UserDimRequest) error {
	logger := middleware.GetLogger(ctx)

	query := fmt.Sprintf(`
		INSERT INTO %s (user_uuid, created_at, country) VALUES (?, ?, ?)
	`, tableUserDim)

	err := c.conn.Exec(ctx, query, req.UserUUID, req.CreatedAt, req.Country)
	if err != nil {
		logger.Error("failed to upsert user dim",
			zap.Error(err),
		)
		return fmt.Errorf("failed to upsert user dim: %w", err)
	}

	logger.Debug("upserted user dim",
		zap.String("country", req.Country),
	)

	return nil
}
