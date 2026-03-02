package di

import (
	"os"
	"strconv"
	"time"

	"libs/consts"
	libsdi "libs/di"

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
	VaultAddr            string
	VaultToken           string
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
	port := os.Getenv("GATEWAY_API_PORT")
	popularityServiceURL := os.Getenv("POPULARITY_SERVICE_URL")
	banditServiceURL := os.Getenv("BANDIT_SERVICE_URL")
	jwtStorePath := os.Getenv("JWT_STORE_PATH")
	applicationName := os.Getenv("VAULT_APPLICATION_NAME")
	jwtTimeoutStr := os.Getenv("VAULT_JWT_TIMEOUT_SECONDS")
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")

	// Validate required environment variables
	if port == "" {
		slogger.Warn("GATEWAY_PORT environment variable is not set")
	}
	if popularityServiceURL == "" {
		slogger.Warn("POPULARITY_SERVICE_URL environment variable is not set")
	}
	if banditServiceURL == "" {
		slogger.Warn("BANDIT_SERVICE_URL environment variable is not set")
	}
	if jwtStorePath == "" {
		slogger.Warn("JWT_STORE_PATH environment variable is not set")
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

	// Set default application name if not provided
	if applicationName == "" {
		applicationName = consts.VaultAppGatewayRecommendation
		slogger.Warnf("VAULT_APPLICATION_NAME environment variable is not set, using default: %s", applicationName)
	}

	return &Config{
		Port:                 port,
		PopularityServiceURL: popularityServiceURL,
		BanditServiceURL:     banditServiceURL,
		JWTStorePath:         jwtStorePath,
		ApplicationName:      applicationName,
		JWTTimeout:           jwtTimeout,
		VaultAddr:            vaultAddr,
		VaultToken:           vaultToken,
		UserUUIDKey:          libsdi.UserUUIDKey,
		RequestIDKey:         libsdi.RequestIDKey,
	}
}

// GetRequestIDKey implements middleware.RequestIDConfig
func (c *Config) GetRequestIDKey() any {
	return c.RequestIDKey
}

// GetUserUUIDKey implements middleware.LoggingConfig and middleware.AuthConfig
func (c *Config) GetUserUUIDKey() (any, bool) {
	return c.UserUUIDKey, true
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
