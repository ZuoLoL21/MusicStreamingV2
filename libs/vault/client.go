package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"libs/consts"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ClientConfig is the interface that services must implement to use vault functionality.
// It requires methods to get the Vault address and token from configuration.
type ClientConfig interface {
	GetVaultAddr() string
	GetVaultToken() string
}

// InitializeKeyVersion fetches the latest key version from Vault Transit and initializes the global version counter.
//
// This should be called during application startup before any JWT operations.
// It queries Vault for the latest version of the Transit key for the given application name.
// The function retries up to 15 times with 3-second delays if Vault is temporarily unavailable.
//
// Returns an error if the key version cannot be determined after all retries.
func InitializeKeyVersion(applicationName string, logger *zap.Logger, config ClientConfig) error {
	latestVersion, err := readKeyVersionDirectHTTP(applicationName, logger, config)
	if err != nil {
		return err
	}

	SetVersion(latestVersion)

	logger.Info("key version initialized successfully",
		zap.String("key_name", applicationName),
		zap.Int32("version", latestVersion))

	return nil
}

// readKeyVersionDirectHTTP reads the Transit key version using raw HTTP.
// This function retries up to 15 times with 3-second delays if Vault is unavailable.
//
// Returns the latest key version or an error if all retries fail.
func readKeyVersionDirectHTTP(applicationName string, logger *zap.Logger, config ClientConfig) (int32, error) {
	maxRetries := 15
	retryDelay := 3 * time.Second

	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		version, err := fetchKeyVersion(applicationName, logger, config)
		if err == nil {
			logger.Info("successfully fetched key version",
				zap.String("key_name", applicationName),
				zap.Int32("version", version),
				zap.Int("attempt", attempt))
			return version, nil
		}

		lastErr = err
		logger.Warn("failed to fetch key version, retrying",
			zap.String("key_name", applicationName),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries),
			zap.Error(err))

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	logger.Error("failed to fetch key version after all retries",
		zap.String("key_name", applicationName),
		zap.Int("attempts", maxRetries),
		zap.Error(lastErr))
	return 0, fmt.Errorf("%s: %w", consts.ErrKeyVersionFetchFailed, lastErr)
}

// fetchKeyVersion makes a single HTTP request to read the Transit key version from Vault.
// It queries the Vault Transit secrets engine at /v1/transit/keys/{applicationName}.
// Returns the latest version of the key or an error if the request fails.
func fetchKeyVersion(applicationName string, logger *zap.Logger, config ClientConfig) (int32, error) {
	vaultAddr := config.GetVaultAddr()
	vaultToken := config.GetVaultToken()

	if vaultAddr == "" || vaultToken == "" {
		return 0, fmt.Errorf(consts.ErrVaultNotConfigured)
	}

	url := fmt.Sprintf("%s/v1/transit/keys/%s", vaultAddr, applicationName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", consts.ErrHTTPRequestFailed, err)
	}

	req.Header.Set("X-Vault-Token", vaultToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf(consts.ErrVaultReturnedError, resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			LatestVersion int32 `json:"latest_version"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("%s: %w", consts.ErrDecodeResponse, err)
	}

	if result.Data.LatestVersion == 0 {
		return 0, fmt.Errorf(consts.ErrLatestVersionMissing)
	}

	return result.Data.LatestVersion, nil
}
