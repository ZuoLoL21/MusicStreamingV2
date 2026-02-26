package di

import (
	"os"

	libsdi "libs/di"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	Port         string
	WarehouseURL string
	TableName    string
	JWTStorePath string
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
	port := os.Getenv("POPULARITY_PORT")
	warehouseURL := os.Getenv("WAREHOUSE_URL")
	tableName := os.Getenv("TABLE_NAME")
	jwtStorePath := os.Getenv("JWT_STORE_PATH")

	// Validate required environment variables
	if port == "" {
		slogger.Warn("POPULARITY_PORT environment variable is not set")
	}
	if warehouseURL == "" {
		slogger.Warn("WAREHOUSE_URL environment variable is not set")
	}
	if tableName == "" {
		slogger.Warn("TABLE_NAME environment variable is not set")
	}
	if jwtStorePath == "" {
		slogger.Warn("JWT_STORE_PATH environment variable is not set")
	}

	return &Config{
		Port:         port,
		WarehouseURL: warehouseURL,
		TableName:    tableName,
		JWTStorePath: jwtStorePath,
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
