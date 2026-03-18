//go:build integration

package integration

import (
	backenddi "backend/internal/di"
	"backend/internal/handlers"
	"backend/tests/integration/builders"
	"context"
	"libs/di"
	"libs/helpers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIntegration_UpdatePassword_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create test user
	userUUID := builders.NewUserBuilder().
		WithEmail("password@test.com").
		WithPassword("OldPassword123!").
		Build(t, ctx, db)

	handler := handlers.NewUserHandler(config, nil, returns, db, nil, nil)

	// Update password request
	updateReq := map[string]string{
		"old_password": "OldPassword123!",
		"new_password": "NewPassword456!",
	}
	req := createJSONRequest(t, "POST", "/users/me/password", updateReq)

	// Wrap with auth
	router := mux.NewRouter()
	router.HandleFunc("/users/me/password", wrapWithAuth(t, handler.UpdatePassword, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "password updated")

	// Verify password was updated in database
	hashedPassword, err := db.GetHashPassword(ctx, userUUID)
	require.NoError(t, err)

	// Old password should no longer work
	require.False(t, helpers.Verify("OldPassword123!", hashedPassword))
	// New password should work
	require.True(t, helpers.Verify("NewPassword456!", hashedPassword))
}

func TestIntegration_UpdatePassword_WrongOldPassword(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create test user
	userUUID := builders.NewUserBuilder().
		WithEmail("password@test.com").
		WithPassword("CorrectPassword123!").
		Build(t, ctx, db)

	handler := handlers.NewUserHandler(config, nil, returns, db, nil, nil)

	// Attempt to update with wrong old password
	updateReq := map[string]string{
		"old_password": "WrongPassword123!",
		"new_password": "NewPassword456!",
	}
	req := createJSONRequest(t, "POST", "/users/me/password", updateReq)

	router := mux.NewRouter()
	router.HandleFunc("/users/me/password", wrapWithAuth(t, handler.UpdatePassword, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 401 Unauthorized
	assertErrorResponse(t, rr, http.StatusUnauthorized, "invalid password")

	// Verify password was NOT changed
	hashedPassword, err := db.GetHashPassword(ctx, userUUID)
	require.NoError(t, err)
	require.True(t, helpers.Verify("CorrectPassword123!", hashedPassword))
	require.False(t, helpers.Verify("NewPassword456!", hashedPassword))
}

func TestIntegration_UpdatePassword_SameAsOld(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create test user
	userUUID := builders.NewUserBuilder().
		WithEmail("password@test.com").
		WithPassword("SamePassword123!").
		Build(t, ctx, db)

	handler := handlers.NewUserHandler(config, nil, returns, db, nil, nil)

	// Update to same password
	updateReq := map[string]string{
		"old_password": "SamePassword123!",
		"new_password": "SamePassword123!", // Same as old
	}
	req := createJSONRequest(t, "POST", "/users/me/password", updateReq)

	router := mux.NewRouter()
	router.HandleFunc("/users/me/password", wrapWithAuth(t, handler.UpdatePassword, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Current implementation allows this - should succeed
	// If business logic changes to forbid, update this test
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestIntegration_UpdatePassword_WeakNewPassword(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create test user
	userUUID := builders.NewUserBuilder().
		WithEmail("password@test.com").
		WithPassword("OldPassword123!").
		Build(t, ctx, db)

	handler := handlers.NewUserHandler(config, nil, returns, db, nil, nil)

	testCases := []struct {
		name        string
		newPassword string
	}{
		{"too_short", "Pass1!"},
		{"only_7_chars", "Pass12!"},
		{"empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			updateReq := map[string]string{
				"old_password": "OldPassword123!",
				"new_password": tc.newPassword,
			}
			req := createJSONRequest(t, "POST", "/users/me/password", updateReq)

			router := mux.NewRouter()
			router.HandleFunc("/users/me/password", wrapWithAuth(t, handler.UpdatePassword, userUUID)).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Should return 400 Bad Request (validation failure)
			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestIntegration_UpdatePassword_Unauthorized(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	logger := zap.NewNop()
	config := &backenddi.Config{}
	vaultConfig := NewTestVaultConfig(t)
	jwtHandler := di.GetJWTHandler(logger, vaultConfig, "service-user-database")
	returns := di.NewReturnManager(logger)

	handler := handlers.NewUserHandler(config, jwtHandler, returns, db, nil, nil)

	// Attempt to update password without authentication
	updateReq := map[string]string{
		"old_password": "OldPassword123!",
		"new_password": "NewPassword456!",
	}
	req := createJSONRequest(t, "POST", "/users/me/password", updateReq)

	// Call handler directly without auth wrapper
	rr := httptest.NewRecorder()
	handler.UpdatePassword(rr, req)

	// Should fail (either 400 or 401) because user UUID is not in context
	assert.True(t, rr.Code == http.StatusBadRequest || rr.Code == http.StatusUnauthorized)
}

func TestIntegration_UpdatePassword_VerifyNewPassword(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	vaultConfig := NewTestVaultConfig(t)
	jwtHandler := di.GetJWTHandler(logger, vaultConfig, "service-user-database")
	returns := di.NewReturnManager(logger)

	// Create test user
	userUUID := builders.NewUserBuilder().
		WithEmail("verify@test.com").
		WithPassword("OldPassword123!").
		Build(t, ctx, db)

	handler := handlers.NewUserHandler(config, jwtHandler, returns, db, nil, nil)

	// Update password
	updateReq := map[string]string{
		"old_password": "OldPassword123!",
		"new_password": "NewPassword456!",
	}
	req := createJSONRequest(t, "POST", "/users/me/password", updateReq)

	router := mux.NewRouter()
	router.HandleFunc("/users/me/password", wrapWithAuth(t, handler.UpdatePassword, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Create AuthHandler for login verification
	authHandler := handlers.NewAuthHandler(config, jwtHandler, returns, db, nil, nil)

	// Verify can login with new password
	loginReq := map[string]string{
		"email":       "verify@test.com",
		"password":    "NewPassword456!",
		"device_id":   "00000000-0000-0000-0000-000000000001",
		"device_name": "test-device",
	}
	loginReqHTTP := createJSONRequest(t, "POST", "/login", loginReq)
	loginRR := httptest.NewRecorder()
	authHandler.Login(loginRR, loginReqHTTP)
	assert.Equal(t, http.StatusOK, loginRR.Code)

	// Verify cannot login with old password
	oldLoginReq := map[string]string{
		"email":       "verify@test.com",
		"password":    "OldPassword123!",
		"device_id":   "00000000-0000-0000-0000-000000000001",
		"device_name": "test-device",
	}
	oldLoginReqHTTP := createJSONRequest(t, "POST", "/login", oldLoginReq)
	oldLoginRR := httptest.NewRecorder()
	authHandler.Login(oldLoginRR, oldLoginReqHTTP)
	assert.Equal(t, http.StatusUnauthorized, oldLoginRR.Code)
}
