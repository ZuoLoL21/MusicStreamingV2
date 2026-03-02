package di

import (
	"libs/vault"

	"go.uber.org/zap"
)

// JWTConfig combines the required configuration interfaces for JWT management
type JWTConfig interface {
	vault.HashicorpConfig
	vault.ClientConfig
}

// GetJWTHandler initializes and returns a fully configured JWT handler using Vault Transit
func GetJWTHandler(logger *zap.Logger, config JWTConfig, applicationName string) *vault.JWTHandler {
	vaultClient, err := vault.NewVaultClient(logger, config)
	if err != nil {
		logger.Fatal("failed to initialize vault client",
			zap.Error(err))
	}

	hashicorpHandler := vault.NewHashicorpHandler(vaultClient, config)

	jwtHandler := vault.NewJWTManager()
	jwtHandler.VaultHandler = hashicorpHandler
	jwtHandler.ApplicationName = applicationName

	logger.Info("JWT handler initialized",
		zap.String("application_name", applicationName))

	return jwtHandler
}
