package client

import (
	"backend/internal/di"
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
			Logger:     logger,
		},
	}
}

func (c *ClickHouseSync) SyncUserDim(userUUID pgtype.UUID, country string, createdAt time.Time) {
	if c == nil {
		return
	}
	go c.syncUserDim(userUUID, country, createdAt)
}

func (c *ClickHouseSync) syncUserDim(userUUID pgtype.UUID, country string, createdAt time.Time) {
	userUUIDStr := uuid.UUID(userUUID.Bytes).String()
	if c.config.EventIngestionServiceURL == "" {
		return
	}

	payload := map[string]any{
		"user_uuid":  userUUIDStr,
		"created_at": createdAt.Format(time.RFC3339),
		"country":    country,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error("failed to marshal user dim payload", zap.Error(err))
		return
	}

	serviceJWT, err := c.jwtHandler.GenerateJwt(
		libsconsts.JWTSubjectService,
		userUUIDStr,
		c.config.JWTExpirationService,
	)
	if err != nil {
		c.logger.Error("failed to generate service JWT", zap.Error(err))
		return
	}

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", "Bearer "+serviceJWT)

	_, statusCode, _, err := c.client.DoProxy(
		context.Background(),
		"POST",
		c.config.EventIngestionServiceURL+"/events/user",
		bytes.NewBuffer(jsonData),
		headers,
		"",
	)

	if err != nil {
		c.logger.Warn("failed to sync user dim to ClickHouse", zap.Error(err))
		return
	}

	if statusCode != http.StatusOK {
		c.logger.Warn("user dim sync failed", zap.Int("status", statusCode))
	} else {
		c.logger.Debug("user dim synced to ClickHouse", zap.String("user_uuid", userUUIDStr))
	}
}
