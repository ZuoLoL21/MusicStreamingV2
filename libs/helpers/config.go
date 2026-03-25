package helpers

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// GetEnvOrDefault returns the environment variable value or a default if not set
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvRequired returns the environment variable value or panics if not set
func GetEnvRequired(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	panic(fmt.Errorf("%s environment variable is not set", key))
}

// ParseDurationMinutes parses a duration string (in minutes) or returns default
func ParseDurationMinutes(value string, defaultDuration time.Duration, logger *zap.SugaredLogger, envVar string) time.Duration {
	if value == "" {
		return defaultDuration
	}
	minutes, err := strconv.Atoi(value)
	if err != nil {
		logger.Errorf("Error parsing %s: %v, using default %v", envVar, err, defaultDuration)
		return defaultDuration
	}
	return time.Minute * time.Duration(minutes)
}

// ParseDurationDays parses a duration string (in days) or returns default
func ParseDurationDays(value string, defaultDuration time.Duration, logger *zap.SugaredLogger, envVar string) time.Duration {
	if value == "" {
		return defaultDuration
	}
	days, err := strconv.Atoi(value)
	if err != nil {
		logger.Errorf("Error parsing %s: %v, using default %v", envVar, err, defaultDuration)
		return defaultDuration
	}
	return time.Hour * 24 * time.Duration(days)
}

// ParseDurationSeconds parses a duration string (in seconds) or returns default
func ParseDurationSeconds(value string, defaultDuration time.Duration, logger *zap.SugaredLogger, envVar string) time.Duration {
	if value == "" {
		return defaultDuration
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		logger.Errorf("Error parsing %s: %v, using default %v", envVar, err, defaultDuration)
		return defaultDuration
	}
	return time.Second * time.Duration(seconds)
}
