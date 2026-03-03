package vault

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HashicorpConfig is the interface that services must implement to use HashicorpHandler
type HashicorpConfig interface {
	GetVaultAddr() string
	GetVaultToken() string
	GetJWTTimeout() time.Duration
}

type HashicorpHandler struct {
	VaultAddr  string
	VaultToken string
	JWTTimeout time.Duration
	HTTPClient *http.Client
}

func NewHashicorpHandler(config HashicorpConfig) *HashicorpHandler {
	return &HashicorpHandler{
		VaultAddr:  config.GetVaultAddr(),
		VaultToken: config.GetVaultToken(),
		JWTTimeout: config.GetJWTTimeout(),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (h *HashicorpHandler) Sign(
	ctx context.Context,
	keyVersion int32,
	applicationName string,
	signingString string) (string, int32, error) {

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), h.JWTTimeout)
		defer cancel()
	}

	payload := map[string]interface{}{
		"input":                base64.StdEncoding.EncodeToString([]byte(signingString)),
		"hash_algorithm":       VaultHashAlgorithm,
		"marshaling_algorithm": VaultMarshalingAlg,
		"key_version":          keyVersion,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/%s/sign/%s", h.VaultAddr, VaultMountPath, applicationName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", h.VaultToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("vault sign returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}

	signature, ok := result.Data[VaultSignatureKey].(string)
	if !ok {
		return "", 0, fmt.Errorf("signature not found in response")
	}

	// Parse
	parts := strings.Split(signature, ":")
	if len(parts) != 3 || parts[0] != VaultSignaturePrefix {
		return "", 0, fmt.Errorf(ErrInvalidFormat)
	}

	version, err := strconv.ParseInt(parts[1][1:], 10, 32)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse version: %w", err)
	}

	return parts[2], int32(version), nil
}

func (h *HashicorpHandler) Verify(
	ctx context.Context,
	keyVersion int32,
	applicationName string,
	signingString string,
	sig []byte) error {

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), h.JWTTimeout)
		defer cancel()
	}

	prependedSignature := fmt.Sprintf(VaultSignatureFormat, keyVersion, sig)

	payload := map[string]interface{}{
		"input":                base64.StdEncoding.EncodeToString([]byte(signingString)),
		"signature":            base64.StdEncoding.EncodeToString([]byte(prependedSignature)),
		"hash_algorithm":       VaultHashAlgorithm,
		"marshaling_algorithm": VaultMarshalingAlg,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/%s/verify/%s", h.VaultAddr, VaultMountPath, applicationName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", h.VaultToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault verify returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	valid, ok := result.Data[VaultValidKey].(bool)
	if !ok {
		return fmt.Errorf("valid field not found in response")
	}

	if !valid {
		return fmt.Errorf(ErrInvalidTransitKey)
	}

	return nil
}
