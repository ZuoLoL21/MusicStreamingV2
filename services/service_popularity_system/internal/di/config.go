package di

import (
	"os"

	libsdi "libs/di"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	WarehouseURL string
	TableName    string
	UserUUIDKey  libsdi.ContextKey
	RequestIDKey libsdi.ContextKey
}

func LoadConfig(logger *zap.Logger) *Config {
	slogger := logger.With(
		zap.String("lifespan", "init"),
	).Sugar()

	err := godotenv.Load()
	if err != nil {
		slogger.Warnf("Error loading .env file: %v", err)
	}

	// Load environment variables
	warehouseURL := os.Getenv("WAREHOUSE_URL")
	tableName := os.Getenv("TABLE_NAME")

	// Validate required environment variables
	if warehouseURL == "" {
		slogger.Warn("WAREHOUSE_URL environment variable is not set")
	}
	if tableName == "" {
		slogger.Warn("TABLE_NAME environment variable is not set")
	}

	return &Config{
		WarehouseURL: warehouseURL,
		TableName:    tableName,
		UserUUIDKey:  libsdi.UserUUIDKey,
		RequestIDKey: libsdi.RequestIDKey,
	}
}

// GetRequestIDKey implements middleware.RequestIDConfig
func (c *Config) GetRequestIDKey() any {
	return c.RequestIDKey
}

// GetUserUUIDKey implements middleware.LoggingConfig
func (c *Config) GetUserUUIDKey() (any, bool) {
	return c.UserUUIDKey, true
}
