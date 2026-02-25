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
	MinIOEndpoint        string
	MinIOAccessKey       string
	MinIOSecretKey       string
	MinIOBucketName      string
	MinIOUseSSL          bool
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
		slogger.Warnf("Error loading .env file: %v", err)
	}

	// Load environment variables
	provider := os.Getenv("USER_CRUD_JWT_PROVIDER_NAME")
	databaseURL := os.Getenv("USER_CRUD_CONNECTION_STRING")
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucketName := os.Getenv("MINIO_BUCKET_NAME")
	minioUseSSLStr := os.Getenv("MINIO_USE_SSL")
	jwtStorePath := os.Getenv("JWT_STORE_PATH")
	subjectNormal := os.Getenv("JWT_SUBJECT_NORMAL")
	jwtTimeNormalStr := os.Getenv("JWT_TIME_IN_M_NORMAL")
	subjectRefresh := os.Getenv("JWT_SUBJECT_REFRESH")
	jwtTimeRefreshStr := os.Getenv("JWT_TIME_IN_D_REFRESH")

	// Validate required environment variables
	if provider == "" {
		slogger.Warn("USER_CRUD_JWT_PROVIDER_NAME environment variable is not set")
	}
	if databaseURL == "" {
		slogger.Warn("USER_CRUD_CONNECTION_STRING environment variable is not set")
	}
	if minioEndpoint == "" {
		slogger.Warn("MINIO_ENDPOINT environment variable is not set")
	}
	if minioAccessKey == "" {
		slogger.Warn("MINIO_ACCESS_KEY environment variable is not set")
	}
	if minioSecretKey == "" {
		slogger.Warn("MINIO_SECRET_KEY environment variable is not set")
	}
	if minioBucketName == "" {
		slogger.Warn("MINIO_BUCKET_NAME environment variable is not set")
	}
	if jwtStorePath == "" {
		slogger.Warn("JWT_STORE_PATH environment variable is not set")
	}
	if subjectNormal == "" {
		slogger.Warn("JWT_SUBJECT_NORMAL environment variable is not set")
	}
	if subjectRefresh == "" {
		slogger.Warn("JWT_SUBJECT_REFRESH environment variable is not set")
	}

	// Parse JWT expiration times
	normalTime := 0
	if jwtTimeNormalStr == "" {
		slogger.Warn("JWT_TIME_IN_M_NORMAL environment variable is not set")
	} else {
		normalTime, err = strconv.Atoi(jwtTimeNormalStr)
		if err != nil {
			slogger.Errorf("Error parsing JWT_TIME_IN_M_NORMAL: %v", err)
		}
	}

	refreshTime := 0
	if jwtTimeRefreshStr == "" {
		slogger.Warn("JWT_TIME_IN_D_REFRESH environment variable is not set")
	} else {
		refreshTime, err = strconv.Atoi(jwtTimeRefreshStr)
		if err != nil {
			slogger.Errorf("Error parsing JWT_TIME_IN_D_REFRESH: %v", err)
		}
	}

	// Parse MinIO SSL setting
	useSSL := false
	if minioUseSSLStr == "true" {
		useSSL = true
	}

	return &Config{
		Provider:             provider,
		DatabaseURL:          databaseURL,
		MinIOEndpoint:        minioEndpoint,
		MinIOAccessKey:       minioAccessKey,
		MinIOSecretKey:       minioSecretKey,
		MinIOBucketName:      minioBucketName,
		MinIOUseSSL:          useSSL,
		JWTStorePath:         jwtStorePath,
		SubjectNormal:        subjectNormal,
		JWTExpirationNormal:  time.Minute * time.Duration(normalTime),
		SubjectRefresh:       subjectRefresh,
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
