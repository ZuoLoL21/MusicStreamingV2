package di

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ClickHouseClient struct {
	conn   driver.Conn
	logger *zap.Logger
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
		conn:   conn,
		logger: logger,
	}, nil
}

func (c *ClickHouseClient) Close() error {
	return c.conn.Close()
}

// InsertListenEvent inserts a music listen event
func (c *ClickHouseClient) InsertListenEvent(ctx context.Context, req ListenEventRequest) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			user_uuid, music_uuid, artist_uuid, album_uuid,
			listen_duration_seconds, track_duration_seconds, completion_ratio
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, tableListenEvents)

	// TODO: Choose one approach
	//var albumUUID interface{}
	//if req.AlbumUUID != nil {
	//	albumUUID = *req.AlbumUUID
	//} else {
	//	albumUUID = nil
	//}
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
		c.logger.Error("failed to insert listen event",
			zap.Error(err),
			zap.String("user_uuid", req.UserUUID.String()),
			zap.String("music_uuid", req.MusicUUID.String()),
		)
		return fmt.Errorf("failed to insert listen event: %w", err)
	}

	c.logger.Debug("inserted listen event",
		zap.String("user_uuid", req.UserUUID.String()),
		zap.String("music_uuid", req.MusicUUID.String()),
	)

	return nil
}

// InsertLikeEvent inserts a music like event
func (c *ClickHouseClient) InsertLikeEvent(ctx context.Context, req LikeEventRequest) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (user_uuid, music_uuid, artist_uuid) VALUES (?, ?, ?)
	`, tableLikeEvents)

	err := c.conn.Exec(ctx, query,
		req.UserUUID,
		req.MusicUUID,
		req.ArtistUUID,
	)
	if err != nil {
		c.logger.Error("failed to insert like event",
			zap.Error(err),
			zap.String("user_uuid", req.UserUUID.String()),
			zap.String("music_uuid", req.MusicUUID.String()),
		)
		return fmt.Errorf("failed to insert like event: %w", err)
	}

	c.logger.Debug("inserted like event",
		zap.String("user_uuid", req.UserUUID.String()),
		zap.String("music_uuid", req.MusicUUID.String()),
	)

	return nil
}

// UpsertTheme inserts/updates a music theme (ReplacingMergeTree)
func (c *ClickHouseClient) UpsertTheme(ctx context.Context, req ThemeEventRequest) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (music_uuid, theme) VALUES (?, ?)
	`, tableTheme)

	err := c.conn.Exec(ctx, query, req.MusicUUID, req.Theme)
	if err != nil {
		c.logger.Error("failed to upsert theme",
			zap.Error(err),
			zap.String("music_uuid", req.MusicUUID.String()),
			zap.String("theme", req.Theme),
		)
		return fmt.Errorf("failed to upsert theme: %w", err)
	}

	c.logger.Debug("upserted theme",
		zap.String("music_uuid", req.MusicUUID.String()),
		zap.String("theme", req.Theme),
	)

	return nil
}

// UpsertUserDim inserts/updates a user dimension (ReplacingMergeTree)
func (c *ClickHouseClient) UpsertUserDim(ctx context.Context, req UserDimRequest) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (user_uuid, created_at, country) VALUES (?, ?, ?)
	`, tableUserDim)

	err := c.conn.Exec(ctx, query, req.UserUUID, req.CreatedAt, req.Country)
	if err != nil {
		c.logger.Error("failed to upsert user dim",
			zap.Error(err),
			zap.String("user_uuid", req.UserUUID.String()),
		)
		return fmt.Errorf("failed to upsert user dim: %w", err)
	}

	c.logger.Debug("upserted user dim",
		zap.String("user_uuid", req.UserUUID.String()),
		zap.String("country", req.Country),
	)

	return nil
}

// Request models
type ListenEventRequest struct {
	UserUUID              uuid.UUID  `json:"user_uuid"`
	MusicUUID             uuid.UUID  `json:"music_uuid"`
	ArtistUUID            uuid.UUID  `json:"artist_uuid"`
	AlbumUUID             *uuid.UUID `json:"album_uuid"`
	ListenDurationSeconds int        `json:"listen_duration_seconds"`
	TrackDurationSeconds  int        `json:"track_duration_seconds"`
	CompletionRatio       float64    `json:"completion_ratio"`
}

type LikeEventRequest struct {
	UserUUID   uuid.UUID `json:"user_uuid"`
	MusicUUID  uuid.UUID `json:"music_uuid"`
	ArtistUUID uuid.UUID `json:"artist_uuid"`
}

type ThemeEventRequest struct {
	MusicUUID uuid.UUID `json:"music_uuid"`
	Theme     string    `json:"theme"`
}

type UserDimRequest struct {
	UserUUID  uuid.UUID `json:"user_uuid"`
	CreatedAt time.Time `json:"created_at"`
	Country   string    `json:"country"`
}
