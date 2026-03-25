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
	Port            string
	WarehouseURL    string
	TableName       string
	JWTStorePath    string
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

	// Load required environment variables (no defaults for connection strings/secrets)
	warehouseURL := helpers.GetEnvRequired("WAREHOUSE_URL")
	vaultAddr := helpers.GetEnvRequired("VAULT_ADDR")
	vaultToken := helpers.GetEnvRequired("VAULT_TOKEN")

	// Load optional environment variables (with sensible defaults)
	port := helpers.GetEnvOrDefault("POPULARITY_PORT", "8003")
	tableName := helpers.GetEnvOrDefault("TABLE_NAME", "popularity_data")
	jwtStorePath := helpers.GetEnvOrDefault("JWT_STORE_PATH", "jwt/popularity")
	applicationName := helpers.GetEnvOrDefault("VAULT_APPLICATION_NAME", consts.VaultAppPopularitySystem)
	jwtTimeout := helpers.ParseDurationSeconds(os.Getenv("VAULT_JWT_TIMEOUT_SECONDS"), consts.JWTTimeoutVault, slogger, "VAULT_JWT_TIMEOUT_SECONDS")

	return &Config{
		Port:            port,
		WarehouseURL:    warehouseURL,
		TableName:       tableName,
		JWTStorePath:    jwtStorePath,
		ApplicationName: applicationName,
		JWTTimeout:      jwtTimeout,
		VaultAddr:       vaultAddr,
		VaultToken:      vaultToken,
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
