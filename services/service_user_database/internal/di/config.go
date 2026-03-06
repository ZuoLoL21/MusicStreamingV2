package di

import (
	"os"
	"strconv"
	"time"

	"libs/consts"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	Provider                 string
	DatabaseURL              string
	MinIOEndpoint            string
	MinIOAccessKey           string
	MinIOSecretKey           string
	MinIOBucketName          string
	EventIngestionServiceURL string
	JWTStorePath             string
	JWTExpirationNormal      time.Duration
	JWTExpirationRefresh     time.Duration
	JWTExpirationService     time.Duration
	ApplicationName          string
	JWTTimeout               time.Duration
	VaultAddr                string
	VaultToken               string
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
	eventIngestionServiceURL := os.Getenv("EVENT_INGESTION_SERVICE_URL")
	jwtStorePath := os.Getenv("JWT_STORE_PATH")
	jwtTimeNormalStr := os.Getenv("JWT_TIME_IN_M_NORMAL")
	jwtTimeRefreshStr := os.Getenv("JWT_TIME_IN_D_REFRESH")
	jwtTimeServiceStr := os.Getenv("JWT_TIME_IN_M_SERVICE")
	applicationName := os.Getenv("VAULT_APPLICATION_NAME")
	jwtTimeoutStr := os.Getenv("VAULT_JWT_TIMEOUT_SECONDS")
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")

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

	// Parse JWT expiration times
	jwtExpirationNormal := consts.JWTExpirationNormal
	if jwtTimeNormalStr != "" {
		normalTime, err := strconv.Atoi(jwtTimeNormalStr)
		if err != nil {
			slogger.Errorf("Error parsing JWT_TIME_IN_M_NORMAL: %v, using default", err)
		} else {
			jwtExpirationNormal = time.Minute * time.Duration(normalTime)
		}
	} else {
		slogger.Warnf("JWT_TIME_IN_M_NORMAL environment variable is not set, using default: %v", consts.JWTExpirationNormal)
	}

	jwtExpirationRefresh := consts.JWTExpirationRefresh
	if jwtTimeRefreshStr != "" {
		refreshTime, err := strconv.Atoi(jwtTimeRefreshStr)
		if err != nil {
			slogger.Errorf("Error parsing JWT_TIME_IN_D_REFRESH: %v, using default", err)
		} else {
			jwtExpirationRefresh = time.Hour * 24 * time.Duration(refreshTime)
		}
	} else {
		slogger.Warnf("JWT_TIME_IN_D_REFRESH environment variable is not set, using default: %v", consts.JWTExpirationRefresh)
	}

	jwtExpirationService := consts.JWTExpirationService
	if jwtTimeServiceStr != "" {
		serviceTime, err := strconv.Atoi(jwtTimeServiceStr)
		if err != nil {
			slogger.Errorf("Error parsing JWT_TIME_IN_D_REFRESH: %v, using default", err)
		} else {
			jwtExpirationService = time.Hour * 24 * time.Duration(serviceTime)
		}
	} else {
		slogger.Warnf("JWT_TIME_IN_D_REFRESH environment variable is not set, using default: %v", consts.JWTExpirationRefresh)
	}

	// Parse JWT timeout for Vault operations
	jwtTimeout := 30 * time.Second
	if jwtTimeoutStr == "" {
		slogger.Warn("VAULT_JWT_TIMEOUT_SECONDS environment variable is not set, using default: 30 seconds")
	} else {
		timeoutSec, err := strconv.Atoi(jwtTimeoutStr)
		if err != nil {
			slogger.Errorf("Error parsing VAULT_JWT_TIMEOUT_SECONDS: %v", err)
		} else {
			jwtTimeout = time.Duration(timeoutSec) * time.Second
		}
	}

	// Set defaults if not provided
	if applicationName == "" {
		applicationName = consts.VaultAppUserDatabase
		slogger.Warnf("VAULT_APPLICATION_NAME environment variable is not set, using default: %s", applicationName)
	}

	return &Config{
		Provider:                 provider,
		DatabaseURL:              databaseURL,
		MinIOEndpoint:            minioEndpoint,
		MinIOAccessKey:           minioAccessKey,
		MinIOSecretKey:           minioSecretKey,
		MinIOBucketName:          minioBucketName,
		EventIngestionServiceURL: eventIngestionServiceURL,
		JWTStorePath:             jwtStorePath,
		JWTExpirationNormal:      jwtExpirationNormal,
		JWTExpirationRefresh:     jwtExpirationRefresh,
		JWTExpirationService:     jwtExpirationService,
		ApplicationName:          applicationName,
		JWTTimeout:               jwtTimeout,
		VaultAddr:                vaultAddr,
		VaultToken:               vaultToken,
	}
}

// GetJWTTimeout implements HashicorpConfig
func (c *Config) GetJWTTimeout() time.Duration {
	return c.JWTTimeout
}

// GetVaultAddr implements VaultConfig
func (c *Config) GetVaultAddr() string {
	return c.VaultAddr
}

// GetVaultToken implements VaultConfig
func (c *Config) GetVaultToken() string {
	return c.VaultToken
}
