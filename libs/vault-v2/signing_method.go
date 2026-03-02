package vault_v2

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
	data := key.(*SigningKey)

	var version = data.Version
	var signature string
	var err error

	for {
		signature, version, err = data.VaultHandler.Sign(nil, data.Version, data.ApplicationName, signingString)

		if err != nil {
			return nil, err
		}
		if version != data.Version {
			data.Version = version
			continue
		}
		break
	}

	return []byte(signature), nil
}

func (h *SigningMethodVault) Verify(signingString string, sig []byte, key interface{}) error {
	data := key.(*SigningKey)

	return data.VaultHandler.Verify(nil, data.Version, data.ApplicationName, signingString, sig)
}
