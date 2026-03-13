package vault

import (
	"fmt"
	"libs/consts"
)

// SigningMethodVault implements the jwt.SigningMethod interface for Vault-based signing.
// It allows the golang-jwt library to use HashiCorp Vault Transit as a signing backend.
type SigningMethodVault struct {
	Algorithm string
}

// Alg returns the algorithm name used for JWT signing.
// This implements the jwt.SigningMethod interface.
func (h *SigningMethodVault) Alg() string {
	return h.Algorithm
}

// SigningKey holds the configuration for signing with Vault Transit.
// It contains references to the Vault handler, application name, and key version.
type SigningKey struct {
	VaultHandler    *HashicorpHandler
	ApplicationName string
	Version         int32
}

// Sign signs the signing string using HashiCorp Vault Transit.
// This implements the jwt.SigningMethod interface.
// The key must be a *SigningKey, otherwise an error is returned.
// It handles key version rotation automatically by detecting version changes during signing.
func (h *SigningMethodVault) Sign(signingString string, key interface{}) ([]byte, error) {
	data, ok := key.(*SigningKey)
	if !ok {
		return nil, fmt.Errorf(consts.ErrInvalidKey)
	}

	var previousVersion = data.Version
	var version int32
	var signature string
	var err error

	for {
		signature, version, err = data.VaultHandler.Sign(nil, previousVersion, data.ApplicationName, signingString)

		if err != nil {
			return nil, err
		}
		if version != previousVersion {
			if !UpdateVersion(version, previousVersion) {
				previousVersion = GetVersion()
			} else {
				previousVersion = version
			}
			continue
		}
		break
	}

	return []byte(signature), nil
}

// Verify verifies the signature using HashiCorp Vault Transit.
// This implements the jwt.SigningMethod interface.
// The key must be a *SigningKey, otherwise an error is returned.
func (h *SigningMethodVault) Verify(signingString string, sig []byte, key interface{}) error {
	data, ok := key.(*SigningKey)
	if !ok {
		return fmt.Errorf(consts.ErrInvalidKey)
	}

	return data.VaultHandler.Verify(nil, data.Version, data.ApplicationName, signingString, sig)
}
