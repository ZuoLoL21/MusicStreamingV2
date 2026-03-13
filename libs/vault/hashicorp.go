package vault

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"libs/consts"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HashicorpConfig is the interface that services must implement to use HashicorpHandler.
// It requires methods to get the Vault address, token, and JWT timeout duration.
type HashicorpConfig interface {
	GetVaultAddr() string
	GetVaultToken() string
	GetJWTTimeout() time.Duration
}

// HashicorpHandler handles communication with HashiCorp Vault for JWT signing operations.
// It provides methods to sign and verify data using Vault's Transit secrets engine.
type HashicorpHandler struct {
	VaultAddr  string
	VaultToken string
	JWTTimeout time.Duration
	HTTPClient *http.Client
}

// NewHashicorpHandler creates a new HashicorpHandler with the given configuration.
// The handler is configured with the Vault address, token, and timeout for JWT operations.
// It uses a default HTTP client with a 30-second timeout.
func NewHashicorpHandler(config HashicorpConfig) *HashicorpHandler {
	return &HashicorpHandler{
		VaultAddr:  config.GetVaultAddr(),
		VaultToken: config.GetVaultToken(),
		JWTTimeout: config.GetJWTTimeout(),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Sign signs data using HashiCorp Vault's Transit secrets engine.
//
// Parameters:
//   - ctx: Context for the request (if nil, a default timeout context is created)
//   - keyVersion: The version of the Transit key to use for signing
//   - applicationName: The name of the Transit key in Vault
//   - signingString: The data to sign
//
// Returns the signature as a string, the key version used, and any error that occurred.
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
		"hash_algorithm":       consts.VaultHashAlgorithm,
		"marshaling_algorithm": consts.VaultMarshalingAlg,
		"key_version":          keyVersion,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("%s: %w", consts.ErrMarshalRequestFailed, err)
	}

	url := fmt.Sprintf("%s/v1/%s/sign/%s", h.VaultAddr, consts.VaultMountPath, applicationName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", 0, fmt.Errorf("%s: %w", consts.ErrCreateRequest, err)
	}

	req.Header.Set("X-Vault-Token", h.VaultToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("%s: %w", consts.ErrHTTPRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf(consts.ErrVaultReturnedError, resp.StatusCode, string(body))
	}

	var result struct {
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("%s: %w", consts.ErrDecodeResponse, err)
	}

	signature, ok := result.Data[consts.VaultSignatureKey].(string)
	if !ok {
		return "", 0, fmt.Errorf(consts.ErrSignatureNotFound)
	}

	parts := strings.Split(signature, ":")
	if len(parts) != 3 || parts[0] != consts.VaultSignaturePrefix {
		return "", 0, fmt.Errorf(consts.ErrInvalidFormat)
	}

	version, err := strconv.ParseInt(parts[1][1:], 10, 32)
	if err != nil {
		return "", 0, fmt.Errorf("%s: %w", consts.ErrParseVersionFailed, err)
	}

	return parts[2], int32(version), nil
}

// Verify verifies a signature using HashiCorp Vault's Transit secrets engine.
//
// Parameters:
//   - ctx: Context for the request (if nil, a default timeout context is created)
//   - keyVersion: The version of the Transit key to use for verification
//   - applicationName: The name of the Transit key in Vault
//   - signingString: The original data that was signed
//   - sig: The signature to verify
//
// Returns nil if verification succeeds, or an error if verification fails.
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

	vaultSignature := fmt.Sprintf(consts.VaultSignatureFormat, keyVersion, string(sig))

	payload := map[string]interface{}{
		"input":                base64.StdEncoding.EncodeToString([]byte(signingString)),
		"signature":            vaultSignature,
		"hash_algorithm":       consts.VaultHashAlgorithm,
		"marshaling_algorithm": consts.VaultMarshalingAlg,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", consts.ErrMarshalRequestFailed, err)
	}

	url := fmt.Sprintf("%s/v1/%s/verify/%s", h.VaultAddr, consts.VaultMountPath, applicationName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("%s: %w", consts.ErrCreateRequest, err)
	}

	req.Header.Set("X-Vault-Token", h.VaultToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", consts.ErrHTTPRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(consts.ErrVaultReturnedError, resp.StatusCode, string(body))
	}

	var result struct {
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("%s: %w", consts.ErrDecodeResponse, err)
	}

	valid, ok := result.Data[consts.VaultValidKey].(bool)
	if !ok {
		return fmt.Errorf(consts.ErrValidFieldNotFound)
	}

	if !valid {
		return fmt.Errorf(consts.ErrInvalidTransitKey)
	}

	return nil
}
