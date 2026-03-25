package di

import (
	"os"
	"time"

	"libs/consts"
	"libs/helpers"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
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

	// Load required environment variables (no defaults for hosts/URLs/secrets)
	databaseURL := helpers.GetEnvRequired("USER_CRUD_CONNECTION_STRING")
	minioEndpoint := helpers.GetEnvRequired("MINIO_ENDPOINT")
	minioAccessKey := helpers.GetEnvRequired("MINIO_ACCESS_KEY")
	minioSecretKey := helpers.GetEnvRequired("MINIO_SECRET_KEY")
	eventIngestionServiceURL := helpers.GetEnvRequired("EVENT_INGESTION_SERVICE_URL")
	vaultAddr := helpers.GetEnvRequired("VAULT_ADDR")
	vaultToken := helpers.GetEnvRequired("VAULT_TOKEN")

	// Load optional environment variables (with sensible defaults)
	minioBucketName := helpers.GetEnvOrDefault("MINIO_BUCKET_NAME", "music-streaming")
	jwtStorePath := helpers.GetEnvOrDefault("JWT_STORE_PATH", "jwt/backend")
	applicationName := helpers.GetEnvOrDefault("VAULT_APPLICATION_NAME", consts.VaultAppUserDatabase)
	jwtExpirationNormal := helpers.ParseDurationMinutes(os.Getenv("JWT_TIME_IN_M_NORMAL"), consts.JWTExpirationNormal, slogger, "JWT_TIME_IN_M_NORMAL")
	jwtExpirationRefresh := helpers.ParseDurationDays(os.Getenv("JWT_TIME_IN_D_REFRESH"), consts.JWTExpirationRefresh, slogger, "JWT_TIME_IN_D_REFRESH")
	jwtExpirationService := helpers.ParseDurationMinutes(os.Getenv("JWT_TIME_IN_M_SERVICE"), consts.JWTExpirationService, slogger, "JWT_TIME_IN_M_SERVICE")
	jwtTimeout := helpers.ParseDurationSeconds(os.Getenv("VAULT_JWT_TIMEOUT_SECONDS"), consts.JWTTimeoutVault, slogger, "VAULT_JWT_TIMEOUT_SECONDS")

	return &Config{
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

// GetVaultAddr implements VaultConfig
func (c *Config) GetVaultAddr() string {
	return c.VaultAddr
}

// GetVaultToken implements VaultConfig
func (c *Config) GetVaultToken() string {
	return c.VaultToken
}
