package vault

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	JWTSubjectNormal  = "normal"
	JWTSubjectRefresh = "refresh"
	JWTSubjectService = "service"
)

type MyCustomClaims struct {
	Uuid string `json:"uuid"`
	jwt.RegisteredClaims
}
type JWTHandler struct {
	VaultHandler    *HashicorpHandler
	ApplicationName string
}

func NewJWTManager() *JWTHandler {
	signingAlg := &SigningMethodVault{HeaderAlgorithm}

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
		jwt.GetSigningMethod(HeaderAlgorithm),
		claims,
	)

	kid := GetVersion()
	t.Header[HeaderKeyID] = kid
	t.Header[HeaderAppName] = h.ApplicationName

	key := SigningKey{
		VaultHandler:    h.VaultHandler,
		ApplicationName: h.ApplicationName,
		Version:         kid,
	}
	s, err := t.SignedString(key)

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
			kid, ok := token.Header[HeaderKeyID]
			if !ok {
				return nil, fmt.Errorf(ErrKIDMissing)
			}

			kidI, err := strconv.ParseInt(kid.(string), 10, 32)
			if err != nil {
				return nil, fmt.Errorf(ErrKIDNotInt, err)
			}

			appName, ok := token.Header[HeaderAppName]
			if !ok {
				return nil, fmt.Errorf(ErrAppNameMissing)
			}

			return SigningKey{
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
			return "", fmt.Errorf(ErrInvalidSubject, subject, claims.Subject)
		}
		return claims.Uuid, nil
	}
	return "", fmt.Errorf(ErrInvalidToken)
}
