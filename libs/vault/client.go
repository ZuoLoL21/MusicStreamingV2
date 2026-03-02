package vault

import (
	"fmt"

	vault "github.com/hashicorp/vault-client-go"
	"go.uber.org/zap"
)

// ClientConfig is the interface that services must implement to use NewVaultClient
type ClientConfig interface {
	GetVaultAddr() string
	GetVaultToken() string
}

// NewVaultClient creates and initializes a Vault client from configuration
func NewVaultClient(logger *zap.Logger, config ClientConfig) (*vault.Client, error) {
	vaultAddr := config.GetVaultAddr()
	vaultToken := config.GetVaultToken()

	if vaultAddr == "" {
		logger.Error("vault address not configured")
		return nil, fmt.Errorf(ErrVaultAddrMissing)
	}

	if vaultToken == "" {
		logger.Error("vault token not configured")
		return nil, fmt.Errorf(ErrVaultTokenMissing)
	}

	client, err := vault.New(
		vault.WithAddress(vaultAddr),
		vault.WithRequestTimeout(30),
	)
	if err != nil {
		logger.Error("failed to create vault client",
			zap.String("vault_addr", vaultAddr),
			zap.Error(err))
		return nil, err
	}

	if err := client.SetToken(vaultToken); err != nil {
		logger.Error("failed to set vault token",
			zap.Error(err))
		return nil, err
	}

	logger.Info("vault client initialized",
		zap.String("vault_addr", vaultAddr))

	return client, nil
}
