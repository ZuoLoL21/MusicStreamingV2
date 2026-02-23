package di

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"libs/vault"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// SecretsManager TODO: add the ability to rotate keys
type SecretsManager struct {
	logger *zap.Logger
	client *api.Client

	// TODO: change this to use REDIS
	privateKeySearch map[string]*ecdsa.PrivateKey
	publicKeySearch  map[string]*ecdsa.PublicKey
}

func GetSecretsManager(logger *zap.Logger) *SecretsManager {
	client := getClient(logger, api.DefaultConfig())
	return &SecretsManager{
		logger:           logger,
		client:           client,
		privateKeySearch: make(map[string]*ecdsa.PrivateKey),
		publicKeySearch:  make(map[string]*ecdsa.PublicKey),
	}
}

func getClient(logger *zap.Logger, config *api.Config) *api.Client {
	client, err := api.NewClient(config)
	if err != nil {
		logger.Fatal("Error creating client", zap.Error(err))
	}
	return client
}

func (h *SecretsManager) GetKeyInfo(path string) (*ecdsa.PublicKey, *ecdsa.PrivateKey, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	keyPair, err := vault.LoadOrCreateKeyPair(ctx, h.client, path)
	if err != nil {
		h.logger.Fatal("Error loading key", zap.String("path", path), zap.Error(err))
	}

	h.publicKeySearch[keyPair.KID] = keyPair.PublicKey
	h.privateKeySearch[keyPair.KID] = keyPair.PrivateKey

	return keyPair.PublicKey, keyPair.PrivateKey, keyPair.KID
}

func (h *SecretsManager) GetPublicKeyFunc() func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return "", fmt.Errorf("kid header missing or invalid")
		}

		key := h.publicKeySearch[kid]
		if key == nil {
			return "", fmt.Errorf("key not found")
		}
		return key, nil
	}
}
