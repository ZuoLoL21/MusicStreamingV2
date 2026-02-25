package di

import (
	"os"

	libsdi "libs/di"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	Port                 string
	PopularityServiceURL string
	BanditServiceURL     string
	RequestIDKey         libsdi.ContextKey
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
	port := os.Getenv("GATEWAY_PORT")
	popularityServiceURL := os.Getenv("POPULARITY_SERVICE_URL")
	banditServiceURL := os.Getenv("BANDIT_SERVICE_URL")

	// Validate required environment variables
	if port == "" {
		slogger.Warn("GATEWAY_PORT environment variable is not set")
	}
	if popularityServiceURL == "" {
		slogger.Warn("POPULARITY_SERVICE_URL environment variable is not set")
	}
	if banditServiceURL == "" {
		slogger.Warn("BANDIT_SERVICE_URL environment variable is not set")
	}

	return &Config{
		Port:                 port,
		PopularityServiceURL: popularityServiceURL,
		BanditServiceURL:     banditServiceURL,
		RequestIDKey:         libsdi.RequestIDKey,
	}
}

// GetRequestIDKey implements middleware.RequestIDConfig
func (c *Config) GetRequestIDKey() any {
	return c.RequestIDKey
}

// GetUserUUIDKey implements middleware.LoggingConfig
func (c *Config) GetUserUUIDKey() (any, bool) {
	return nil, false // No authentication yet
}
