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
	Port                     string
	UserDatabaseServiceURL   string
	RecommendationServiceURL string
	EventIngestionServiceURL string
	JWTStorePath             string
	JWTExpirationService     time.Duration
	ApplicationName          string
	JWTTimeout               time.Duration
	VaultAddr                string
	VaultToken               string
	UserUUIDKey              libsdi.ContextKey
	ServiceJWTKey            libsdi.ContextKey
	RequestIDKey             libsdi.ContextKey
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
	userDatabaseServiceURL := os.Getenv("USER_DATABASE_SERVICE_URL")
	recommendationServiceURL := os.Getenv("RECOMMENDATION_SERVICE_URL")
	eventIngestionServiceURL := os.Getenv("EVENT_INGESTION_SERVICE_URL")
	jwtStorePath := os.Getenv("JWT_STORE_PATH")
	jwtTimeServiceStr := os.Getenv("JWT_TIME_IN_M_SERVICE")
	applicationName := os.Getenv("VAULT_APPLICATION_NAME")
	jwtTimeoutStr := os.Getenv("VAULT_JWT_TIMEOUT_SECONDS")
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")

	// Validate required environment variables
	if port == "" {
		slogger.Warn("GATEWAY_API_PORT environment variable is not set")
	}
	if userDatabaseServiceURL == "" {
		slogger.Warn("USER_DATABASE_SERVICE_URL environment variable is not set")
	}
	if recommendationServiceURL == "" {
		slogger.Warn("RECOMMENDATION_SERVICE_URL environment variable is not set")
	}
	if eventIngestionServiceURL == "" {
		slogger.Warn("EVENT_INGESTION_SERVICE_URL environment variable is not set")
	}
	if jwtStorePath == "" {
		slogger.Warn("JWT_STORE_PATH environment variable is not set")
	}

	// Parse JWT expiration time for service JWT
	serviceTime := 2
	if jwtTimeServiceStr == "" {
		slogger.Warn("JWT_TIME_IN_M_SERVICE environment variable is not set, using default: 2 minutes")
	} else {
		serviceTime, err = strconv.Atoi(jwtTimeServiceStr)
		if err != nil {
			slogger.Errorf("Error parsing JWT_TIME_IN_M_SERVICE: %v", err)
			serviceTime = 2
		}
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
		applicationName = consts.VaultAppGatewayAPI
		slogger.Warnf("VAULT_APPLICATION_NAME environment variable is not set, using default: %s", applicationName)
	}

	return &Config{
		Port:                     port,
		UserDatabaseServiceURL:   userDatabaseServiceURL,
		RecommendationServiceURL: recommendationServiceURL,
		EventIngestionServiceURL: eventIngestionServiceURL,
		JWTStorePath:             jwtStorePath,
		JWTExpirationService:     time.Minute * time.Duration(serviceTime),
		ApplicationName:          applicationName,
		JWTTimeout:               jwtTimeout,
		VaultAddr:                vaultAddr,
		VaultToken:               vaultToken,
		UserUUIDKey:              libsdi.UserUUIDKey,
		ServiceJWTKey:            libsdi.ServiceJWTKey,
		RequestIDKey:             libsdi.RequestIDKey,
	}
}

// GetRequestIDKey implements middleware.RequestIDConfig
func (c *Config) GetRequestIDKey() any {
	return c.RequestIDKey
}

// GetUserUUIDKey implements middleware.AuthConfig
func (c *Config) GetUserUUIDKey() (any, bool) {
	return c.UserUUIDKey, true
}

// GetServiceJWTKey implements middleware.ServiceJWTConfig
func (c *Config) GetServiceJWTKey() (any, bool) {
	return c.ServiceJWTKey, true
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
