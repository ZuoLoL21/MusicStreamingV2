package vault

import "fmt"

type SigningMethodVault struct {
	Algorithm string
}

func (h *SigningMethodVault) Alg() string {
	return h.Algorithm
}

type SigningKey struct {
	VaultHandler    *HashicorpHandler
	ApplicationName string
	Version         int32
}

func (h *SigningMethodVault) Sign(signingString string, key interface{}) ([]byte, error) {
	data, ok := key.(*SigningKey)
	if !ok {
		return nil, fmt.Errorf(ErrInvalidKey)
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

func (h *SigningMethodVault) Verify(signingString string, sig []byte, key interface{}) error {
	data, ok := key.(*SigningKey)
	if !ok {
		return fmt.Errorf(ErrInvalidKey)
	}

	return data.VaultHandler.Verify(nil, data.Version, data.ApplicationName, signingString, sig)
}
