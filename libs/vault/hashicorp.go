package vault

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	vault "github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

// HashicorpConfig is the interface that services must implement to use HashicorpHandler (e.g. Config must have these methods)
type HashicorpConfig interface {
	GetJWTTimeout() time.Duration
}
type HashicorpHandler struct {
	Client     *vault.Client
	JWTTimeout time.Duration
}

func NewHashicorpHandler(c *vault.Client, config HashicorpConfig) *HashicorpHandler {
	return &HashicorpHandler{Client: c, JWTTimeout: config.GetJWTTimeout()}
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

	resp, err := h.Client.Secrets.TransitSign(
		ctx,
		applicationName,
		schema.TransitSignRequest{
			Input:               base64.StdEncoding.EncodeToString([]byte(signingString)),
			HashAlgorithm:       VaultHashAlgorithm,
			MarshalingAlgorithm: VaultMarshalingAlg,
			KeyVersion:          keyVersion,
		},
		vault.WithMountPath(VaultMountPath),
	)
	if err != nil {
		return "", 0, err
	}

	signature, ok := resp.Data[VaultSignatureKey].(string)
	if !ok {
		panic(resp.Data)
	}

	parts := strings.Split(signature, ":")
	if len(parts) != 3 || parts[0] != VaultSignaturePrefix {
		panic(ErrInvalidFormat)
	}

	version, err := strconv.ParseInt(parts[1][1:], 10, 32)
	if err != nil {
		panic(err)
	}
	signature = parts[2]

	return signature, int32(version), nil
}

func (h *HashicorpHandler) Verify(ctx context.Context, keyVersion int32, applicationName string, signingString string, sig []byte) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), h.JWTTimeout)
		defer cancel()
	}

	prependedSignature := fmt.Sprintf(VaultSignatureFormat, keyVersion, sig)

	resp, err := h.Client.Secrets.TransitVerify(
		ctx,
		applicationName,
		schema.TransitVerifyRequest{
			Input:               base64.StdEncoding.EncodeToString([]byte(signingString)),
			Signature:           base64.StdEncoding.EncodeToString([]byte(prependedSignature)),
			HashAlgorithm:       VaultHashAlgorithm,
			MarshalingAlgorithm: VaultMarshalingAlg,
		},
		vault.WithMountPath(VaultMountPath),
	)
	if err != nil {
		return err
	}

	valid, ok := resp.Data[VaultValidKey].(bool)
	if !ok {
		panic(resp.Data)
	}

	if !valid {
		return fmt.Errorf(ErrInvalidTransitKey)
	}

	return nil
}
