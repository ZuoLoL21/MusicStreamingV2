package di

import (
	"os"
	"path/filepath"

	libsdi "libs/di"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	StorageLocation string
	ListenAddr      string
	RequestIDKey    libsdi.ContextKey
}

func LoadConfig(logger *zap.Logger) *Config {
	slogger := logger.With(
		zap.String("lifespan", "init"),
	).Sugar()

	err := godotenv.Load()
	if err != nil {
		slogger.Errorf("Error loading .env file: %v", err)
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "127.0.0.1:8000"
	}

	return &Config{
		StorageLocation: filepath.Clean(os.Getenv("DATA_LOCATION")),
		ListenAddr:      listenAddr,
		RequestIDKey:    libsdi.RequestIDKey,
	}
}

func (c *Config) GetRequestIDKey() any {
	return c.RequestIDKey
}

func (c *Config) GetUserUUIDKey() (any, bool) {
	return nil, false
}
