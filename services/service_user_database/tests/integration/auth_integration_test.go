//go:build integration

package integration

import (
	"backend/tests/integration/builders"
	"context"
	"libs/di"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIntegration_Auth_JWTGeneration_Service(t *testing.T) {
	config := NewTestVaultConfig(t)
	logger := zap.NewNop()

	// Initialize JWT handler for service JWTs
	jwtHandler := di.GetJWTHandler(logger, config, "service-user-database")

	// Create test user UUID
	testUUID := "550e8400-e29b-41d4-a716-446655440000"

	// Generate service JWT
	serviceJWT, err := jwtHandler.GenerateJwt("service", testUUID, 2*time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, serviceJWT, "service JWT should be generated")

	// Validate the JWT
	extractedUUID, err := jwtHandler.ValidateJwt("service", serviceJWT)
	require.NoError(t, err, "service JWT should be valid")
	assert.Equal(t, testUUID, extractedUUID, "extracted UUID should match")
}

func TestIntegration_Auth_JWTValidation_ExpiredToken(t *testing.T) {
	config := NewTestVaultConfig(t)
	logger := zap.NewNop()
	jwtHandler := di.GetJWTHandler(logger, config, "service-user-database")

	testUUID := "550e8400-e29b-41d4-a716-446655440000"

	// Generate JWT with very short expiration
	serviceJWT, err := jwtHandler.GenerateJwt("service", testUUID, 1*time.Millisecond)
	require.NoError(t, err)
	require.NotEmpty(t, serviceJWT)

	// Wait for token to expire
	time.Sleep(5 * time.Millisecond)

	// Attempt to validate expired JWT
	_, err := jwtHandler.ValidateJwt("service", serviceJWT)
	assert.Error(t, err, "expired JWT should fail validation")
}

func TestIntegration_Auth_JWTValidation_WrongSubject(t *testing.T) {
	config := NewTestVaultConfig(t)
	logger := zap.NewNop()
	jwtHandler := di.GetJWTHandler(logger, config, "service-user-database")

	testUUID := "550e8400-e29b-41d4-a716-446655440000"

	// Generate JWT with "service" subject
	serviceJWT, err := jwtHandler.GenerateJwt("service", testUUID, 2*time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, serviceJWT)

	// Attempt to validate with wrong subject (e.g., "normal")
	_, err := jwtHandler.ValidateJwt("normal", serviceJWT)
	assert.Error(t, err, "JWT with wrong subject should fail validation")
}

func TestIntegration_Auth_JWTValidation_MalformedToken(t *testing.T) {
	config := NewTestVaultConfig(t)
	logger := zap.NewNop()
	jwtHandler := di.GetJWTHandler(logger, config, "service-user-database")

	// Attempt to validate malformed JWT
	_, err := jwtHandler.ValidateJwt("service", "invalid.jwt.token")
	assert.Error(t, err, "malformed JWT should fail validation")
}

func TestIntegration_Auth_UserLoginFlow(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	config := NewTestVaultConfig(t)
	ctx := context.Background()
	logger := zap.NewNop()

	// Create a test user in database
	userUUID := builders.NewUserBuilder().
		WithEmail("auth-test@example.com").
		WithPassword("TestPassword123!").
		Build(t, ctx, db)

	// Simulate login: Get user by email (this would happen in login handler)
	user, err := db.GetUserByEmail(ctx, "auth-test@example.com")
	require.NoError(t, err)
	require.Equal(t, userUUID.Bytes, user.Uuid.Bytes)

	// Generate user JWT (normally done after password verification)
	jwtHandler := di.GetJWTHandler(logger, config, "gateway-api")
	userJWT, err := jwtHandler.GenerateJwt("normal", builders.UUIDToString(user.Uuid), 10*time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, userJWT)

	// Validate user JWT
	extractedUUID, err := jwtHandler.ValidateJwt("normal", userJWT)
	require.NoError(t, err)
	assert.Equal(t, builders.UUIDToString(user.Uuid), extractedUUID)

	// Simulate gateway transforming user JWT to service JWT
	serviceJWTHandler := di.GetJWTHandler(logger, config, "service-user-database")
	serviceJWT, err := serviceJWTHandler.GenerateJwt("service", extractedUUID, 2*time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, serviceJWT)

	// Validate service JWT at backend
	backendExtractedUUID, err := serviceJWTHandler.ValidateJwt("service", serviceJWT)
	require.NoError(t, err)
	assert.Equal(t, extractedUUID, backendExtractedUUID)
}

func TestIntegration_Auth_MultipleApplications(t *testing.T) {
	config := NewTestVaultConfig(t)
	logger := zap.NewNop()
	testUUID := "550e8400-e29b-41d4-a716-446655440000"

	// Create JWT handlers for different applications
	gatewayHandler := di.GetJWTHandler(logger, config, "gateway-api")
	backendHandler := di.GetJWTHandler(logger, config, "service-user-database")

	// Generate JWT from gateway
	gatewayJWT, err := gatewayHandler.GenerateJwt("normal", testUUID, 10*time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, gatewayJWT)

	// Validate with correct application
	extractedUUID, err := gatewayHandler.ValidateJwt("normal", gatewayJWT)
	require.NoError(t, err)
	assert.Equal(t, testUUID, extractedUUID)

	// Generate service JWT from backend
	backendJWT, err := backendHandler.GenerateJwt("service", testUUID, 2*time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, backendJWT)

	// Validate with correct application
	extractedUUID, err = backendHandler.ValidateJwt("service", backendJWT)
	require.NoError(t, err)
	assert.Equal(t, testUUID, extractedUUID)
}
