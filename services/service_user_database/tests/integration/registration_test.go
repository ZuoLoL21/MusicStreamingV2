//go:build integration

package integration

import (
	backenddi "backend/internal/di"
	"backend/internal/handlers"
	"context"
	"libs/di"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestIntegration_Register_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	vaultConfig := NewTestVaultConfig(t)
	jwtHandler := di.GetJWTHandler(logger, vaultConfig, "service-user-database")
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	handler := handlers.NewAuthHandler(config, jwtHandler, returns, db, fileStorage, nil)

	// Create registration request with valid data
	formFields := map[string]string{
		"username":    "validuser123",
		"email":       "validuser@test.com",
		"password":    "SecurePass123!",
		"country":     "US",
		"bio":         "Test bio",
		"device_id":   "00000000-0000-0000-0000-000000000001",
		"device_name": "test-device",
	}
	req := createMultipartRequest(t, "POST", "/register", "", "", nil, formFields)

	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	// Should return 201 Created with JWT tokens
	var response map[string]interface{}
	assertJSONResponse(t, rr, http.StatusCreated, &response)
	assert.Contains(t, response, "access_token")
	assert.Contains(t, response, "refresh_token")
	assert.Contains(t, response, "user_uuid")

	// Verify user was created in database
	user, err := db.GetUserByEmail(ctx, "validuser@test.com")
	assert.NoError(t, err)
	assert.Equal(t, "validuser123", user.Username)
	assert.Equal(t, "US", user.Country)
}

func TestIntegration_Register_DuplicateEmail(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	vaultConfig := NewTestVaultConfig(t)
	jwtHandler := di.GetJWTHandler(logger, vaultConfig, "service-user-database")
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	handler := handlers.NewAuthHandler(config, jwtHandler, returns, db, fileStorage, nil)

	// Create first user
	formFields1 := map[string]string{
		"username":    "user1",
		"email":       "duplicate@test.com",
		"password":    "Password123!",
		"country":     "US",
		"device_id":   "00000000-0000-0000-0000-000000000001",
		"device_name": "test-device",
	}
	req1 := createMultipartRequest(t, "POST", "/register", "", "", nil, formFields1)
	rr1 := httptest.NewRecorder()
	handler.Register(rr1, req1)
	assert.Equal(t, http.StatusCreated, rr1.Code)

	// Attempt to register with same email
	formFields2 := map[string]string{
		"username":    "user2",
		"email":       "duplicate@test.com", // Same email
		"password":    "Password456!",
		"country":     "CA",
		"device_id":   "00000000-0000-0000-0000-000000000002",
		"device_name": "test-device",
	}
	req2 := createMultipartRequest(t, "POST", "/register", "", "", nil, formFields2)
	rr2 := httptest.NewRecorder()
	handler.Register(rr2, req2)

	// Should return 409 Conflict
	assertErrorResponse(t, rr2, http.StatusConflict, "email already in use")

	// Verify only one user exists
	users, err := db.GetUserByEmail(ctx, "duplicate@test.com")
	assert.NoError(t, err)
	assert.Equal(t, "user1", users.Username) // First user should exist
}

func TestIntegration_Register_InvalidEmail(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	logger := zap.NewNop()
	config := &backenddi.Config{}
	vaultConfig := NewTestVaultConfig(t)
	jwtHandler := di.GetJWTHandler(logger, vaultConfig, "service-user-database")
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	handler := handlers.NewAuthHandler(config, jwtHandler, returns, db, fileStorage, nil)

	testCases := []struct {
		name  string
		email string
	}{
		{"empty_email", ""},
		{"missing_at", "notanemail.com"},
		{"missing_domain", "user@"},
		{"missing_local", "@domain.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formFields := map[string]string{
				"username":    "testuser",
				"email":       tc.email,
				"password":    "Password123!",
				"country":     "US",
				"device_id":   "00000000-0000-0000-0000-000000000001",
				"device_name": "test-device",
			}
			req := createMultipartRequest(t, "POST", "/register", "", "", nil, formFields)
			rr := httptest.NewRecorder()
			handler.Register(rr, req)

			// Should return 400 Bad Request
			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestIntegration_Register_WeakPassword(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	logger := zap.NewNop()
	config := &backenddi.Config{}
	vaultConfig := NewTestVaultConfig(t)
	jwtHandler := di.GetJWTHandler(logger, vaultConfig, "service-user-database")
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	handler := handlers.NewAuthHandler(config, jwtHandler, returns, db, fileStorage, nil)

	testCases := []struct {
		name     string
		password string
	}{
		{"too_short", "Pass1!"},
		{"empty", ""},
		{"only_7_chars", "Pass12!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formFields := map[string]string{
				"username":    "testuser",
				"email":       "test@example.com",
				"password":    tc.password,
				"country":     "US",
				"device_id":   "00000000-0000-0000-0000-000000000001",
				"device_name": "test-device",
			}
			req := createMultipartRequest(t, "POST", "/register", "", "", nil, formFields)
			rr := httptest.NewRecorder()
			handler.Register(rr, req)

			// Should return 400 Bad Request
			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestIntegration_Register_ShortUsername(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	logger := zap.NewNop()
	config := &backenddi.Config{}
	vaultConfig := NewTestVaultConfig(t)
	jwtHandler := di.GetJWTHandler(logger, vaultConfig, "service-user-database")
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	handler := handlers.NewAuthHandler(config, jwtHandler, returns, db, fileStorage, nil)

	testCases := []struct {
		name     string
		username string
	}{
		{"4_chars", "user"},
		{"3_chars", "usr"},
		{"1_char", "u"},
		{"empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formFields := map[string]string{
				"username":    tc.username,
				"email":       "test@example.com",
				"password":    "Password123!",
				"country":     "US",
				"device_id":   "00000000-0000-0000-0000-000000000001",
				"device_name": "test-device",
			}
			req := createMultipartRequest(t, "POST", "/register", "", "", nil, formFields)
			rr := httptest.NewRecorder()
			handler.Register(rr, req)

			// Should return 400 Bad Request
			assertErrorResponse(t, rr, http.StatusBadRequest, "username required")
		})
	}
}

func TestIntegration_Register_MissingFields(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	logger := zap.NewNop()
	config := &backenddi.Config{}
	vaultConfig := NewTestVaultConfig(t)
	jwtHandler := di.GetJWTHandler(logger, vaultConfig, "service-user-database")
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	handler := handlers.NewAuthHandler(config, jwtHandler, returns, db, fileStorage, nil)

	testCases := []struct {
		name        string
		formFields  map[string]string
		expectedMsg string
	}{
		{
			name: "missing_email",
			formFields: map[string]string{
				"username":    "testuser",
				"password":    "Password123!",
				"country":     "US",
				"device_id":   "00000000-0000-0000-0000-000000000001",
				"device_name": "test-device",
			},
			expectedMsg: "email required",
		},
		{
			name: "missing_username",
			formFields: map[string]string{
				"email":       "test@example.com",
				"password":    "Password123!",
				"country":     "US",
				"device_id":   "00000000-0000-0000-0000-000000000001",
				"device_name": "test-device",
			},
			expectedMsg: "username required",
		},
		{
			name: "missing_password",
			formFields: map[string]string{
				"username":    "testuser",
				"email":       "test@example.com",
				"country":     "US",
				"device_id":   "00000000-0000-0000-0000-000000000001",
				"device_name": "test-device",
			},
			expectedMsg: "password required",
		},
		{
			name: "missing_country",
			formFields: map[string]string{
				"username":    "testuser",
				"email":       "test@example.com",
				"password":    "Password123!",
				"device_id":   "00000000-0000-0000-0000-000000000001",
				"device_name": "test-device",
			},
			expectedMsg: "country required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createMultipartRequest(t, "POST", "/register", "", "", nil, tc.formFields)
			rr := httptest.NewRecorder()
			handler.Register(rr, req)

			// Should return 400 Bad Request
			assertErrorResponse(t, rr, http.StatusBadRequest, tc.expectedMsg)
		})
	}
}

func TestIntegration_Register_EmptyStrings(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	logger := zap.NewNop()
	config := &backenddi.Config{}
	vaultConfig := NewTestVaultConfig(t)
	jwtHandler := di.GetJWTHandler(logger, vaultConfig, "service-user-database")
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	handler := handlers.NewAuthHandler(config, jwtHandler, returns, db, fileStorage, nil)

	testCases := []struct {
		name       string
		formFields map[string]string
	}{
		{
			name: "empty_username",
			formFields: map[string]string{
				"username":    "",
				"email":       "test@example.com",
				"password":    "Password123!",
				"country":     "US",
				"device_id":   "00000000-0000-0000-0000-000000000001",
				"device_name": "test-device",
			},
		},
		{
			name: "empty_email",
			formFields: map[string]string{
				"username":    "testuser",
				"email":       "",
				"password":    "Password123!",
				"country":     "US",
				"device_id":   "00000000-0000-0000-0000-000000000001",
				"device_name": "test-device",
			},
		},
		{
			name: "empty_password",
			formFields: map[string]string{
				"username":    "testuser",
				"email":       "test@example.com",
				"password":    "",
				"country":     "US",
				"device_id":   "00000000-0000-0000-0000-000000000001",
				"device_name": "test-device",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createMultipartRequest(t, "POST", "/register", "", "", nil, tc.formFields)
			rr := httptest.NewRecorder()
			handler.Register(rr, req)

			// Should return 400 Bad Request
			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestIntegration_Register_VeryLongStrings(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	logger := zap.NewNop()
	config := &backenddi.Config{}
	vaultConfig := NewTestVaultConfig(t)
	jwtHandler := di.GetJWTHandler(logger, vaultConfig, "service-user-database")
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	handler := handlers.NewAuthHandler(config, jwtHandler, returns, db, fileStorage, nil)

	// Create a very long string (>1000 chars)
	longString := string(make([]byte, 1001))
	for i := range longString {
		longString = longString[:i] + "a" + longString[i+1:]
	}

	formFields := map[string]string{
		"username":    longString,
		"email":       "test@example.com",
		"password":    "Password123!",
		"country":     "US",
		"device_id":   "00000000-0000-0000-0000-000000000001",
		"device_name": "test-device",
	}
	req := createMultipartRequest(t, "POST", "/register", "", "", nil, formFields)
	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	// Server may reject very long strings (either 400 or 500)
	// We just verify it doesn't create the user
	assert.NotEqual(t, http.StatusCreated, rr.Code)
}
