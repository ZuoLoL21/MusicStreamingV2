package client

import (
	"backend/internal/di"
	sqlhandler "backend/sql/sqlc"
	"bytes"
	"context"
	"encoding/json"
	libsclients "libs/clients"
	libsconsts "libs/consts"
	libsdi "libs/di"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type ClickHouseSync struct {
	logger     *zap.Logger
	config     *di.Config
	jwtHandler *libsdi.JWTHandler
	client     *libsclients.BaseClient
}

func NewClickHouseSync(logger *zap.Logger, config *di.Config, jwtHandler *libsdi.JWTHandler) *ClickHouseSync {
	return &ClickHouseSync{
		logger:     logger,
		config:     config,
		jwtHandler: jwtHandler,
		client: &libsclients.BaseClient{
			HttpClient: &http.Client{Timeout: 5 * time.Second},
		},
	}
}

func (c *ClickHouseSync) SyncUserDim(userUUID pgtype.UUID, deviceID, country string, createdAt time.Time) {
	if c == nil {
		return
	}
	go c.syncUserDim(userUUID, deviceID, country, createdAt)
}

func (c *ClickHouseSync) syncUserDim(userUUID pgtype.UUID, deviceID, country string, createdAt time.Time) {
	if c.config.EventIngestionServiceURL == "" {
		return
	}

	userUUIDStr := uuid.UUID(userUUID.Bytes).String()
	payload := map[string]any{
		"user_uuid":  userUUIDStr,
		"created_at": createdAt.Format(time.RFC3339),
		"country":    country,
	}

	if err := c.sendEvent("/events/user", payload, userUUIDStr, deviceID); err != nil {
		c.logger.Warn("failed to sync user dim to ClickHouse", zap.Error(err))
	} else {
		c.logger.Debug("user dim synced to ClickHouse", zap.String("user_uuid", userUUIDStr))
	}
}

func (c *ClickHouseSync) SyncListenEvent(userUUID, musicUUID pgtype.UUID, deviceID string, music sqlhandler.Music, listenDurationSeconds *int32, completionPercentage *float64) {
	if c == nil {
		return
	}
	go c.syncListenEvent(userUUID, musicUUID, deviceID, music, listenDurationSeconds, completionPercentage)
}

func (c *ClickHouseSync) syncListenEvent(userUUID, musicUUID pgtype.UUID, deviceID string, music sqlhandler.Music, listenDurationSeconds *int32, completionPercentage *float64) {
	if c.config.EventIngestionServiceURL == "" {
		return
	}

	userUUIDStr := uuid.UUID(userUUID.Bytes).String()

	var completionRatio float64
	if completionPercentage != nil {
		completionRatio = *completionPercentage / 100.0
	}

	payload := map[string]any{
		"user_uuid":               userUUIDStr,
		"music_uuid":              uuid.UUID(musicUUID.Bytes).String(),
		"artist_uuid":             uuid.UUID(music.FromArtist.Bytes).String(),
		"listen_duration_seconds": listenDurationSeconds,
		"track_duration_seconds":  music.DurationSeconds,
		"completion_ratio":        completionRatio,
	}
	if music.InAlbum.Valid {
		payload["album_uuid"] = uuid.UUID(music.InAlbum.Bytes).String()
	}

	if err := c.sendEvent("/events/listen", payload, userUUIDStr, deviceID); err != nil {
		c.logger.Warn("failed to sync listen event to ClickHouse", zap.Error(err))
	} else {
		c.logger.Debug("listen event synced to ClickHouse", zap.String("user_uuid", userUUIDStr), zap.String("music_uuid", uuid.UUID(musicUUID.Bytes).String()))
	}
}

// sendEvent is a helper that generates a service JWT and sends an event payload to the ingestion service
func (c *ClickHouseSync) sendEvent(path string, payload map[string]any, userUUID, deviceID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serviceJWT, err := c.jwtHandler.GenerateJwtWithDevice(
		libsconsts.JWTSubjectService,
		userUUID,
		deviceID,
		c.config.JWTExpirationService,
	)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", "Bearer "+serviceJWT)

	_, statusCode, _, err := c.client.DoProxy(
		ctx,
		"POST",
		c.config.EventIngestionServiceURL+path,
		bytes.NewBuffer(jsonData),
		headers,
		"",
	)

	if err != nil {
		return err
	}

	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return nil
	}

	return nil
}
