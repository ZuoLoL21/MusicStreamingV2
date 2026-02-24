package di

import (
	libsdi "libs/di"
	"os"

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
		slogger.Errorf("Error loading .env file: %v", err)
	}

	return &Config{
		WarehouseURL: os.Getenv("WAREHOUSE_URL"),
		TableName:    os.Getenv("TABLE_NAME"),
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
