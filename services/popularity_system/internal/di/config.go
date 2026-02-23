package di

import (
	libsdi "libs/di"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
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
		UserUUIDKey:  libsdi.UserUUIDKey,
		RequestIDKey: libsdi.RequestIDKey,
	}
}

func (c *Config) GetRequestIDKey() any {
	return c.RequestIDKey
}

func (c *Config) GetUserUUIDKey() (any, bool) {
	return c.UserUUIDKey, true
}
