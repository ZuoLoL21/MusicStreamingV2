package di

import (
	"fmt"
	"libs/vault"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type MyCustomClaims struct {
	Uuid string `json:"uuid"`
	jwt.RegisteredClaims
}

// JWTConfig combines the required configuration interfaces for JWT management
type JWTConfig interface {
	vault.HashicorpConfig
	vault.ClientConfig
}

// GetJWTHandler initializes and returns a fully configured JWT handler using Vault Transit
func GetJWTHandler(logger *zap.Logger, config JWTConfig, applicationName string) *JWTHandler {
	if err := vault.InitializeKeyVersion(applicationName, logger, config); err != nil {
		logger.Fatal("failed to initialize key version",
			zap.String("application_name", applicationName),
			zap.Error(err))
	}

	hashicorpHandler := vault.NewHashicorpHandler(config)

	jwtHandler := NewJWTManager()
	jwtHandler.VaultHandler = hashicorpHandler
	jwtHandler.ApplicationName = applicationName

	logger.Info("JWT handler initialized",
		zap.String("application_name", applicationName))

	return jwtHandler
}

type JWTHandler struct {
	VaultHandler    *vault.HashicorpHandler
	ApplicationName string
}

func NewJWTManager() *JWTHandler {
	signingAlg := &vault.SigningMethodVault{Algorithm: vault.HeaderAlgorithm}

	jwt.RegisterSigningMethod(
		signingAlg.Alg(),
		func() jwt.SigningMethod { return signingAlg },
	)
	return &JWTHandler{}
}

func (h *JWTHandler) GenerateJwt(subject string, uuid string, duration time.Duration) string {
	claims := MyCustomClaims{
		Uuid: uuid,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}
	t := jwt.NewWithClaims(
		jwt.GetSigningMethod(vault.HeaderAlgorithm),
		claims,
	)

	kid := vault.GetVersion()
	t.Header[vault.HeaderKeyID] = kid
	t.Header[vault.HeaderAppName] = h.ApplicationName

	key := vault.SigningKey{
		VaultHandler:    h.VaultHandler,
		ApplicationName: h.ApplicationName,
		Version:         kid,
	}
	s, err := t.SignedString(&key)

	if err != nil {
		panic(err)
	}
	return s
}

func (h *JWTHandler) ValidateJwt(
	subject string,
	tokenStr string,
) (string, error) {
	claims := &MyCustomClaims{}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			kid, ok := token.Header[vault.HeaderKeyID]
			if !ok {
				return nil, fmt.Errorf(vault.ErrKIDMissing)
			}

			var kidI int64
			if kidStr, ok := kid.(string); ok {
				var err error
				kidI, err = strconv.ParseInt(kidStr, 10, 32)
				if err != nil {
					return nil, fmt.Errorf(vault.ErrKIDNotInt, err)
				}
			} else {
				kidI = int64(kid.(float64))
			}

			appName, ok := token.Header[vault.HeaderAppName]
			if !ok {
				return nil, fmt.Errorf(vault.ErrAppNameMissing)
			}

			return &vault.SigningKey{
				VaultHandler:    h.VaultHandler,
				ApplicationName: appName.(string),
				Version:         int32(kidI),
			}, nil
		},
	)
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		if claims.Subject != subject {
			return "", fmt.Errorf(vault.ErrInvalidSubject, subject, claims.Subject)
		}
		return claims.Uuid, nil
	}
	return "", fmt.Errorf(vault.ErrInvalidToken)
}
