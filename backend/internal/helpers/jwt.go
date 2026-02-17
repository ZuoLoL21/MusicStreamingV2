package helpers

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type MyCustomClaims struct {
	Uuid string `json:"uuid"`
	Kid  string `json:"kid"`
	jwt.StandardClaims
}

func GenerateJwt(subject string, uuid string, key *ecdsa.PrivateKey, kid string, duration time.Duration) string {
	claims := MyCustomClaims{
		Uuid: uuid,
		StandardClaims: jwt.StandardClaims{
			Subject:   subject,
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(duration).Unix(),
		},
	}
	t := jwt.NewWithClaims(
		jwt.SigningMethodES256,
		claims,
	)
	t.Header["kid"] = kid
	s, err := t.SignedString(key)

	if err != nil {
		panic(err)
	}
	return s
}

func ValidateJwt(
	subject string,
	tokenStr string,
	keyGetter func(token *jwt.Token) (interface{}, error),
) (string, error) {
	claims := &MyCustomClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, keyGetter)
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		if claims.Subject != subject {
			return "", fmt.Errorf("invalid subject - required %v, current %v", subject, claims.Subject)
		}
		return claims.Uuid, nil
	}
	return "", fmt.Errorf("invalid token")
}
