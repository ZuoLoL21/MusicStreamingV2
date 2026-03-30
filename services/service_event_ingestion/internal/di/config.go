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
	Port                       string
	ClickHouseConnectionString string

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
	clickHouseConnectionString := helpers.GetEnvRequired("CLICKHOUSE_CONNECTION_STRING")
	vaultAddr := helpers.GetEnvRequired("VAULT_ADDR")
	vaultToken := helpers.GetEnvRequired("VAULT_TOKEN")

	// Optional environment variables
	port := helpers.GetEnvOrDefault("EVENT_INGESTION_PORT", "8080")
	jwtTimeout := helpers.ParseDurationSeconds(os.Getenv("VAULT_JWT_TIMEOUT_SECONDS"), consts.JWTTimeoutVault, slogger, "VAULT_JWT_TIMEOUT_SECONDS")

	return &Config{
		Port:                       port,
		ClickHouseConnectionString: clickHouseConnectionString,
		ApplicationName:            consts.VaultAppEventIngestion,
		JWTTimeout:                 jwtTimeout,
		VaultAddr:                  vaultAddr,
		VaultToken:                 vaultToken,
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
