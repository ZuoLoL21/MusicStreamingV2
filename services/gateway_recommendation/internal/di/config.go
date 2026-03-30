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
	Port                 string
	PopularityServiceURL string
	BanditServiceURL     string

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
	popularityServiceURL := helpers.GetEnvRequired("POPULARITY_SERVICE_URL")
	banditServiceURL := helpers.GetEnvRequired("BANDIT_SERVICE_URL")
	vaultAddr := helpers.GetEnvRequired("VAULT_ADDR")
	vaultToken := helpers.GetEnvRequired("VAULT_TOKEN")

	// Optional environment variables
	port := helpers.GetEnvOrDefault("GATEWAY_RECOMMENDATION_PORT", "8080")
	jwtExpirationService := helpers.ParseDurationMinutes(os.Getenv("JWT_TIME_IN_M_SERVICE"), consts.JWTExpirationService, slogger, "JWT_TIME_IN_M_SERVICE")
	jwtTimeout := helpers.ParseDurationSeconds(os.Getenv("VAULT_JWT_TIMEOUT_SECONDS"), consts.JWTTimeoutVault, slogger, "VAULT_JWT_TIMEOUT_SECONDS")

	return &Config{
		Port:                 port,
		PopularityServiceURL: popularityServiceURL,
		BanditServiceURL:     banditServiceURL,
		JWTExpirationService: jwtExpirationService,
		ApplicationName:      consts.VaultAppGatewayRecommendation,
		JWTTimeout:           jwtTimeout,
		VaultAddr:            vaultAddr,
		VaultToken:           vaultToken,
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
