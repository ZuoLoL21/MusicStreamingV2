package vault

import (
	"context"
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

// InitializeKeyVersion fetches the latest key version from Vault Transit and initializes the global version counter
func InitializeKeyVersion(client *vault.Client, applicationName string, logger *zap.Logger) error {
	ctx := context.Background()

	resp, err := client.Read(ctx, fmt.Sprintf("transit/keys/%s", applicationName))
	if err != nil {
		logger.Error("failed to read key information from vault",
			zap.String("key_name", applicationName),
			zap.Error(err))
		return fmt.Errorf("failed to read key info: %w", err)
	}

	keysData, ok := resp.Data["keys"]
	if !ok {
		logger.Error("keys not found in vault key response",
			zap.String("key_name", applicationName))
		return fmt.Errorf("keys not found in response")
	}

	keysMap, ok := keysData.(map[string]interface{})
	if !ok {
		logger.Error("keys is not a map",
			zap.String("key_name", applicationName),
			zap.Any("keys_data", keysData))
		return fmt.Errorf("keys is not a map: %T", keysData)
	}

	var latestVersion int32
	for versionStr := range keysMap {
		var versionNum int32
		if _, err := fmt.Sscanf(versionStr, "%d", &versionNum); err != nil {
			logger.Warn("failed to parse version number",
				zap.String("version_str", versionStr),
				zap.Error(err))
			continue
		}
		if versionNum > latestVersion {
			latestVersion = versionNum
		}
	}

	if latestVersion == 0 {
		logger.Error("no valid version found in keys",
			zap.String("key_name", applicationName))
		return fmt.Errorf("no valid version found in keys")
	}

	SetVersion(latestVersion)

	logger.Info("key version initialized",
		zap.String("key_name", applicationName),
		zap.Int32("version", latestVersion))

	return nil
}
