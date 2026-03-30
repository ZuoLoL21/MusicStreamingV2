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
	// Service-specific configuration
	Port                     string
	UserDatabaseServiceURL   string
	RecommendationServiceURL string
	EventIngestionServiceURL string

	// JWT configuration
	JWTExpirationService time.Duration

	// Vault configuration
	ApplicationName string
	JWTTimeout      time.Duration
	VaultAddr       string
	VaultToken      string
}

func LoadConfig(logger *zap.Logger) *Config {
	slogger := logger.With(
		zap.String("lifespan", "init"),
	).Sugar()

	err := godotenv.Load()
	if err != nil {
		slogger.Warnf("Error loading .env file: %v", err)
	}

	// Required environment variables
	userDatabaseServiceURL := helpers.GetEnvRequired("USER_DATABASE_SERVICE_URL")
	recommendationServiceURL := helpers.GetEnvRequired("RECOMMENDATION_SERVICE_URL")
	eventIngestionServiceURL := helpers.GetEnvRequired("EVENT_INGESTION_SERVICE_URL")
	vaultAddr := helpers.GetEnvRequired("VAULT_ADDR")
	vaultToken := helpers.GetEnvRequired("VAULT_TOKEN")

	// Optional environment variables
	port := helpers.GetEnvOrDefault("GATEWAY_API_PORT", "8080")
	jwtExpirationService := helpers.ParseDurationMinutes(os.Getenv("JWT_TIME_IN_M_SERVICE"), consts.JWTExpirationService, slogger, "JWT_TIME_IN_M_SERVICE")
	jwtTimeout := helpers.ParseDurationSeconds(os.Getenv("VAULT_JWT_TIMEOUT_SECONDS"), consts.JWTTimeoutVault, slogger, "VAULT_JWT_TIMEOUT_SECONDS")

	return &Config{
		Port:                     port,
		UserDatabaseServiceURL:   userDatabaseServiceURL,
		RecommendationServiceURL: recommendationServiceURL,
		EventIngestionServiceURL: eventIngestionServiceURL,
		JWTExpirationService:     jwtExpirationService,
		ApplicationName:          consts.VaultAppGatewayAPI,
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
