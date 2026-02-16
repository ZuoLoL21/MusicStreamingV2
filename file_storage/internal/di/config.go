package di

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	StorageLocation string
	RequestIDKey    string
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
		StorageLocation: filepath.Clean(os.Getenv("DATA_LOCATION")),
		RequestIDKey:    "request_id",
	}
}
