package di

import (
	"os"
	"strconv"
	"time"

	libsdi "libs/di"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	Provider             string
	DatabaseURL          string
	JWTStorePath         string
	SubjectNormal        string
	JWTExpirationNormal  time.Duration
	SubjectRefresh       string
	JWTExpirationRefresh time.Duration
	UserUUIDKey          libsdi.ContextKey
	RequestIDKey         libsdi.ContextKey
}

func LoadConfig(logger *zap.Logger) *Config {
	slogger := logger.With(
		zap.String("lifespan", "init"),
	).Sugar()

	err := godotenv.Load()
	if err != nil {
		slogger.Errorf("Error loading .env file: %v", err)
	}

	normalTime, err := strconv.Atoi(os.Getenv("JWT_TIME_IN_M_NORMAL"))
	if err != nil {
		slogger.Errorf("Error parsing TIME_IN_M_NORMAL: %v", err)
	}

	refreshTime, err := strconv.Atoi(os.Getenv("JWT_TIME_IN_D_REFRESH"))
	if err != nil {
		slogger.Errorf("Error parsing TIME_IN_D_REFRESH: %v", err)
	}

	return &Config{
		Provider:             os.Getenv("USER_CRUD_JWT_PROVIDER_NAME"),
		DatabaseURL:          os.Getenv("USER_CRUD_CONNECTION_STRING"),
		JWTStorePath:         os.Getenv("JWT_STORE_PATH"),
		SubjectNormal:        os.Getenv("JWT_SUBJECT_NORMAL"),
		JWTExpirationNormal:  time.Minute * time.Duration(normalTime),
		SubjectRefresh:       os.Getenv("JWT_SUBJECT_REFRESH"),
		JWTExpirationRefresh: time.Hour * 24 * time.Duration(refreshTime),
		UserUUIDKey:          libsdi.UserUUIDKey,
		RequestIDKey:         libsdi.RequestIDKey,
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
