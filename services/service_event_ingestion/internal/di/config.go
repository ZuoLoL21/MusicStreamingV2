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
	ClickHouseConnectionString string
	JWTStorePath               string
	JWTTimeout                 time.Duration
	ApplicationName            string
	VaultAddr                  string
	VaultToken                 string
}

func LoadConfig(logger *zap.Logger) *Config {
	slogger := logger.With(
		zap.String("lifespan", "init"),
	).Sugar()

	err := godotenv.Load()
	if err != nil {
		slogger.Warnf("Error loading .env file: %v", err)
	}

	// Load required environment variables (no defaults for connection strings/secrets)
	clickHouseConnectionString := helpers.GetEnvRequired("CLICKHOUSE_CONNECTION_STRING")
	vaultAddr := helpers.GetEnvRequired("VAULT_ADDR")
	vaultToken := helpers.GetEnvRequired("VAULT_TOKEN")

	// Load optional environment variables (with sensible defaults)
	jwtStorePath := helpers.GetEnvOrDefault("JWT_STORE_PATH", "jwt/event-ingestion")
	applicationName := helpers.GetEnvOrDefault("VAULT_APPLICATION_NAME", consts.VaultAppEventIngestion)
	jwtTimeout := helpers.ParseDurationSeconds(os.Getenv("VAULT_JWT_TIMEOUT_SECONDS"), consts.JWTTimeoutVault, slogger, "VAULT_JWT_TIMEOUT_SECONDS")

	return &Config{
		ClickHouseConnectionString: clickHouseConnectionString,
		JWTStorePath:               jwtStorePath,
		JWTTimeout:                 jwtTimeout,
		ApplicationName:            applicationName,
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
