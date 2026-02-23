package di

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// ContextKey is an unexported named type used as context keys to avoid
// collisions with plain string keys from other packages.
type ContextKey string

type Config struct {
	StorageLocation string
	ListenAddr      string
	RequestIDKey    ContextKey
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
		RequestIDKey:    "request_id",
	}
}
