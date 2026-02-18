package di

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type ContextKey string

type Config struct {
	Provider             string
	DatabaseURL          string
	JWTStorePath         string
	SubjectNormal        string
	JWTExpirationNormal  time.Duration
	SubjectRefresh       string
	JWTExpirationRefresh time.Duration
	UserUUIDKey          ContextKey
	RequestIDKey         ContextKey
}

func LoadConfig(logger *zap.Logger) *Config {
	slogger := logger.With(
		zap.String("lifespan", "init"),
	).Sugar()

	err := godotenv.Load()
	if err != nil {
		slogger.Errorf("Error loading .env file: %v", err)
	}

	normalTime, err := strconv.Atoi(os.Getenv("TIME_IN_M_NORMAL"))
	if err != nil {
		slogger.Errorf("Error parsing TIME_IN_M_NORMAL: %v", err)
	}

	refreshTime, err := strconv.Atoi(os.Getenv("TIME_IN_D_REFRESH"))
	if err != nil {
		slogger.Errorf("Error parsing TIME_IN_D_REFRESH: %v", err)
	}

	return &Config{
		Provider:             os.Getenv("PROVIDER"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		JWTStorePath:         os.Getenv("JWT_STORE_PATH"),
		SubjectNormal:        os.Getenv("SUBJECT_NORMAL"),
		JWTExpirationNormal:  time.Minute * time.Duration(normalTime),
		SubjectRefresh:       os.Getenv("SUBJECT_REFRESH"),
		JWTExpirationRefresh: time.Hour * 24 * time.Duration(refreshTime),
		UserUUIDKey:          ContextKey("userUuid"),
		RequestIDKey:         ContextKey("requestId"),
	}
}
