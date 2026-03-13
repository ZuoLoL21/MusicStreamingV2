package di

import (
	"fmt"
	"libs/consts"
	"libs/vault"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// MyCustomClaims represents the custom JWT claims used in the music streaming application.
// It includes a Uuid field for the user's unique identifier along with the standard
// JWT registered claims (Subject, IssuedAt, ExpiresAt).
type MyCustomClaims struct {
	Uuid string `json:"uuid"`
	jwt.RegisteredClaims
}

// JWTConfig is an interface that combines vault.HashicorpConfig and vault.ClientConfig.
// Services must implement both interfaces to use JWT functionality.
type JWTConfig interface {
	vault.HashicorpConfig
	vault.ClientConfig
}

// GetJWTHandler creates and initializes a new JWTHandler with the given configuration.
// It takes a logger, configuration, and application name as parameters.
//
// The function initializes the Vault key version, creates a Hashicorp handler,
// and returns a fully configured JWTHandler ready for token generation and validation.
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

// JWTHandler handles JWT token generation and validation using HashiCorp Vault as the signing backend.
// It contains a reference to the Vault handler and the application name for token signing.
type JWTHandler struct {
	VaultHandler    *vault.HashicorpHandler
	ApplicationName string
}

// NewJWTManager creates and initializes a new JWTHandler.
// It registers the Vault signing method with the JWT library and returns
// an empty JWTHandler. The VaultHandler and ApplicationName must be set
// before using the handler for token operations.
func NewJWTManager() *JWTHandler {
	signingAlg := &vault.SigningMethodVault{Algorithm: consts.HeaderAlgorithm}

	jwt.RegisterSigningMethod(
		signingAlg.Alg(),
		func() jwt.SigningMethod { return signingAlg },
	)
	return &JWTHandler{}
}

// GenerateJwt generates a new JWT token with the given subject, UUID, and expiration duration.
//
//   - The subject identifies the token type (e.g., "normal", "refresh", "service").
//   - The UUID is stored in the token claims for user identification.
//   - The duration determines when the token expires.
//
// Returns the signed token string and any error that occurred.
func (h *JWTHandler) GenerateJwt(subject, uuid string, duration time.Duration) (string, error) {
	claims := MyCustomClaims{
		Uuid: uuid,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}
	token := jwt.NewWithClaims(
		jwt.GetSigningMethod(consts.HeaderAlgorithm),
		claims,
	)

	keyID := vault.GetVersion()
	token.Header[consts.HeaderKeyID] = keyID
	token.Header[consts.HeaderAppName] = h.ApplicationName

	signingKey := vault.SigningKey{
		VaultHandler:    h.VaultHandler,
		ApplicationName: h.ApplicationName,
		Version:         keyID,
	}
	tokenString, err := token.SignedString(&signingKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ValidateJwt validates a JWT token and returns the UUID from the claims if valid.
// It takes the expected subject and the token string as parameters.
//
// Returns the UUID from the token claims if validation succeeds.
// Returns an error if the token is invalid, expired, or has an unexpected subject.
func (h *JWTHandler) ValidateJwt(
	subject string,
	tokenStr string,
) (string, error) {
	claims := &MyCustomClaims{}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			kid, ok := token.Header[consts.HeaderKeyID]
			if !ok {
				return nil, fmt.Errorf(consts.ErrKIDMissing)
			}

			var kidI int64
			if kidStr, ok := kid.(string); ok {
				var err error
				kidI, err = strconv.ParseInt(kidStr, 10, 32)
				if err != nil {
					return nil, fmt.Errorf(consts.ErrKIDNotInt, err)
				}
			} else {
				kidI = int64(kid.(float64))
			}

			appName, ok := token.Header[consts.HeaderAppName]
			if !ok {
				return nil, fmt.Errorf(consts.ErrAppNameMissing)
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
			return "", fmt.Errorf(consts.ErrInvalidSubject, subject, claims.Subject)
		}
		return claims.Uuid, nil
	}
	return "", fmt.Errorf(consts.ErrInvalidToken)
}
