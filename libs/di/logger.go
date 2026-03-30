package di

import (
	"os"

	"go.uber.org/zap"
)

// InitLogger creates a production logger with service name and environment metadata.
// The environment defaults to "development" if not set.
func InitLogger(serviceName string) *zap.Logger {
	logger, _ := zap.NewProduction()
	logger = logger.WithOptions(zap.AddCaller())

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	logger = logger.With(
		zap.String("service_name", serviceName),
		zap.String("environment", environment),
	)

	return logger
}
