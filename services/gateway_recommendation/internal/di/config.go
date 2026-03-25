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
	Port                 string
	PopularityServiceURL string
	BanditServiceURL     string
	JWTStorePath         string
	ApplicationName      string
	JWTTimeout           time.Duration
	JWTExpirationService time.Duration
	VaultAddr            string
	VaultToken           string
}

func LoadConfig(logger *zap.Logger) *Config {
	slogger := logger.With(
		zap.String("lifespan", "init"),
	).Sugar()

	err := godotenv.Load()
	if err != nil {
		slogger.Warnf("Error loading .env file: %v", err)
	}

	// Load required environment variables (no defaults for URLs/secrets)
	popularityServiceURL := helpers.GetEnvRequired("POPULARITY_SERVICE_URL")
	banditServiceURL := helpers.GetEnvRequired("BANDIT_SERVICE_URL")
	vaultAddr := helpers.GetEnvRequired("VAULT_ADDR")
	vaultToken := helpers.GetEnvRequired("VAULT_TOKEN")

	// Load optional environment variables (with sensible defaults)
	port := helpers.GetEnvOrDefault("GATEWAY_RECOMMENDATION_PORT", "8002")
	jwtStorePath := helpers.GetEnvOrDefault("JWT_STORE_PATH", "jwt/gateway-recommendation")
	applicationName := helpers.GetEnvOrDefault("VAULT_APPLICATION_NAME", consts.VaultAppGatewayRecommendation)
	jwtExpirationService := helpers.ParseDurationMinutes(os.Getenv("JWT_TIME_IN_M_SERVICE"), consts.JWTExpirationService, slogger, "JWT_TIME_IN_M_SERVICE")
	jwtTimeout := helpers.ParseDurationSeconds(os.Getenv("VAULT_JWT_TIMEOUT_SECONDS"), consts.JWTTimeoutVault, slogger, "VAULT_JWT_TIMEOUT_SECONDS")

	return &Config{
		Port:                 port,
		PopularityServiceURL: popularityServiceURL,
		BanditServiceURL:     banditServiceURL,
		JWTStorePath:         jwtStorePath,
		ApplicationName:      applicationName,
		JWTTimeout:           jwtTimeout,
		JWTExpirationService: jwtExpirationService,
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
