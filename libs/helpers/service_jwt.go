package helpers

import (
	"crypto/ecdsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const ServiceJWTSubject = "service"

// GenerateServiceJwt generates a service-to-service JWT token
// This is a convenience wrapper around GenerateJwt with the "service" subject
func GenerateServiceJwt(uuid string, key *ecdsa.PrivateKey, kid string, duration time.Duration) string {
	return GenerateJwt(ServiceJWTSubject, uuid, key, kid, duration)
}

// ValidateServiceJwt validates a service JWT and returns the embedded user UUID
// This is a convenience wrapper around ValidateJwt with the "service" subject
func ValidateServiceJwt(tokenStr string, keyGetter func(*jwt.Token) (interface{}, error)) (string, error) {
	return ValidateJwt(ServiceJWTSubject, tokenStr, keyGetter)
}
