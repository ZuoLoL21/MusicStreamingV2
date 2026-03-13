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

	// ClickHouse configuration
	clickHouseConnectionString := os.Getenv("CLICKHOUSE_CONNECTION_STRING")
	if clickHouseConnectionString == "" {
		logger.Fatal("CLICKHOUSE_CONNECTION_STRING is required")
	}

	// JWT and Vault configuration
	jwtStorePath := os.Getenv("JWT_STORE_PATH")
	if jwtStorePath == "" {
		slogger.Warn("JWT_STORE_PATH environment variable is not set")
	}
	jwtTimeoutStr := os.Getenv("VAULT_JWT_TIMEOUT_SECONDS")
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

	applicationName := os.Getenv("VAULT_APPLICATION_NAME")
	if applicationName == "" {
		applicationName = consts.VaultAppEventIngestion
		slogger.Warnf("VAULT_APPLICATION_NAME environment variable is not set, using default: %s", applicationName)
	}

	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")

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
