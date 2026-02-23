package di

import (
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type ContextKey string

type Config struct {
	UserUUIDKey  ContextKey
	RequestIDKey ContextKey
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
		UserUUIDKey:  ContextKey("userUuid"),
		RequestIDKey: ContextKey("requestId"),
	}
}
