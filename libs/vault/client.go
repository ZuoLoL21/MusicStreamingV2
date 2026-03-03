package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ClientConfig is the interface that services must implement to use NewVaultClient
type ClientConfig interface {
	GetVaultAddr() string
	GetVaultToken() string
}

// InitializeKeyVersion fetches the latest key version from Vault Transit and initializes the global version counter
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

// readKeyVersionDirectHTTP reads the Transit key version using raw HTTP (bypassing buggy vault-client-go)
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
	return 0, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// fetchKeyVersion makes a single HTTP request to read the Transit key
func fetchKeyVersion(applicationName string, logger *zap.Logger, config ClientConfig) (int32, error) {
	vaultAddr := config.GetVaultAddr()
	vaultToken := config.GetVaultToken()

	if vaultAddr == "" || vaultToken == "" {
		return 0, fmt.Errorf("vault address or token not configured")
	}

	url := fmt.Sprintf("%s/v1/transit/keys/%s", vaultAddr, applicationName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
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
		return 0, fmt.Errorf("vault returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			LatestVersion int32 `json:"latest_version"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Data.LatestVersion == 0 {
		return 0, fmt.Errorf("latest_version is 0 or missing")
	}

	return result.Data.LatestVersion, nil
}
