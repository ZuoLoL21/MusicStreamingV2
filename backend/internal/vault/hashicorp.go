package vault

import (
	"context"
	"crypto/ecdsa"
	"errors"

	vault "github.com/hashicorp/vault/api"
)

type KeyPair struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	KID        string
}

const (
	vaultPrivateKeyField = "private_key_pem"
	vaultPublicKeyField  = "public_key_pem"
	vaultKIDField        = "kid"
)

func LoadOrCreateKeyPair(
	ctx context.Context,
	client *vault.Client,
	secretPath string,
) (*KeyPair, error) {
	key, err := LoadKeyPair(ctx, client, secretPath)
	if err == nil {
		return key, nil
	}

	if !errors.Is(err, vault.ErrSecretNotFound) {
		return nil, err
	}

	key, err = generateKeyPair()
	if err != nil {
		return nil, err
	}

	err = storeKeyPair(ctx, client, secretPath, key)
	if err != nil {
		return LoadKeyPair(ctx, client, secretPath)
	}

	return key, nil
}

func LoadKeyPair(
	ctx context.Context,
	client *vault.Client,
	secretPath string,
) (*KeyPair, error) {

	secret, err := client.KVv2("secret").Get(ctx, secretPath)
	if err != nil {
		return nil, err
	}

	data := secret.Data

	priPEM, ok := data[vaultPrivateKeyField].(string)
	if !ok {
		return nil, errors.New("invalid private key format in vault")
	}

	pubPEM, ok := data[vaultPublicKeyField].(string)
	if !ok {
		return nil, errors.New("invalid public key format in vault")
	}

	kid, ok := data[vaultKIDField].(string)
	if !ok {
		return nil, errors.New("missing kid in vault")
	}

	privateKey, err := parsePrivateKeyPEM([]byte(priPEM))
	if err != nil {
		return nil, err
	}

	publicKey, err := parsePublicKeyPEM([]byte(pubPEM))
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		KID:        kid,
	}, nil
}
