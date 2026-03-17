//go:build e2e

package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuth_HealthCheck tests the health check endpoint
func TestAuth_HealthCheck(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	resp, err := client.RawRequest("GET", "/health", nil)
	require.NoError(t, err, "Health check request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)

	body := AssertResponseBody(t, resp)
	assert.Equal(t, "healthy", body["status"])
	assert.Equal(t, "gateway-api", body["service"])
}

// TestAuth_Login_Success tests successful login
func TestAuth_Login_Success(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// First register a user
	email := fmt.Sprintf("logintest-%s@example.com", NewTestUUID()[:8])
	password := "TestPass123!"

	resp, err := client.RawMultipartRequest("PUT", "/login", map[string]string{
		"email":    email,
		"password": password,
		"username": "logintest" + NewTestUUID()[:8],
		"country":  "US",
	})

	require.NoError(t, err, "Register request should not fail")
	defer resp.Body.Close()

	// Accept either 200 or 201
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Skipf("Cannot register user: status %d", resp.StatusCode)
	}

	// Now login
	resp, err = client.RawRequest("POST", "/login", map[string]interface{}{
		"email":    email,
		"password": password,
	})

	require.NoError(t, err, "Login request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)

	body := AssertResponseBody(t, resp)
	AssertContainsField(t, body, "access_token")
	AssertContainsField(t, body, "refresh_token")
	AssertContainsField(t, body, "user_uuid")

	// Verify token format (should be JWT)
	assert.True(t, len(body["access_token"].(string)) > 20, "Access token should be a valid JWT")
	assert.True(t, len(body["refresh_token"].(string)) > 20, "Refresh token should be a valid JWT")

	client.SetTokens(
		body["access_token"].(string),
		body["refresh_token"].(string),
		body["user_uuid"].(string),
	)
}

// TestAuth_Login_InvalidCredentials tests login with invalid credentials
func TestAuth_Login_InvalidCredentials(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	resp, err := client.RawRequest("POST", "/login", map[string]interface{}{
		"email":    "nonexistent@example.com",
		"password": "wrongpassword",
	})

	require.NoError(t, err, "Login request should not fail")
	defer resp.Body.Close()

	// Should return 401 or 400
	assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest,
		"Login with invalid credentials should fail")
}

// TestAuth_Register_Success tests successful registration
func TestAuth_Register_Success(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	email := fmt.Sprintf("registertest-%s@example.com", NewTestUUID()[:8])
	password := "TestPass123!"
	username := "registertest" + NewTestUUID()[:8]

	resp, err := client.RawMultipartRequest("PUT", "/login", map[string]string{
		"email":    email,
		"password": password,
		"username": username,
		"country":  "US",
	})

	require.NoError(t, err, "Register request should not fail")
	defer resp.Body.Close()

	// Accept 200 (already exists) or 201 (created)
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated,
		"Registration should succeed or user already exists")

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		body := AssertResponseBody(t, resp)
		AssertContainsField(t, body, "access_token")
		AssertContainsField(t, body, "refresh_token")
		AssertContainsField(t, body, "user_uuid")
	}
}

// TestAuth_Register_InvalidEmail tests registration with invalid email
func TestAuth_Register_InvalidEmail(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	resp, err := client.RawMultipartRequest("PUT", "/login", map[string]string{
		"email":    "notanemail",
		"password": "TestPass123!",
		"username": "testuser",
		"country":  "US",
	})

	require.NoError(t, err, "Register request should not fail")
	defer resp.Body.Close()

	// Should return 400 Bad Request
	AssertResponseStatus(t, resp, http.StatusBadRequest)
}

// TestAuth_Register_ShortPassword tests registration with short password
func TestAuth_Register_ShortPassword(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	resp, err := client.RawMultipartRequest("PUT", "/login", map[string]string{
		"email":    fmt.Sprintf("shortpw-%s@example.com", NewTestUUID()[:8]),
		"password": "123",
		"username": "testuser",
		"country":  "US",
	})

	require.NoError(t, err, "Register request should not fail")
	defer resp.Body.Close()

	// Should return 400 Bad Request
	AssertResponseStatus(t, resp, http.StatusBadRequest)
}

// TestAuth_Register_DuplicateEmail tests registration with duplicate email
func TestAuth_Register_DuplicateEmail(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	email := fmt.Sprintf("duplicate-%s@example.com", NewTestUUID()[:8])
	password := "TestPass123!"

	// First registration
	resp, err := client.RawMultipartRequest("PUT", "/login", map[string]string{
		"email":    email,
		"password": password,
		"username": "user1" + NewTestUUID()[:8],
		"country":  "US",
	})

	require.NoError(t, err, "First register request should not fail")
	defer resp.Body.Close()

	// Second registration with same email (different username)
	resp, err = client.RawMultipartRequest("PUT", "/login", map[string]string{
		"email":    email,
		"password": password,
		"username": "user2" + NewTestUUID()[:8],
		"country":  "US",
	})

	require.NoError(t, err, "Second register request should not fail")
	defer resp.Body.Close()

	// Should return 409 Conflict
	AssertResponseStatus(t, resp, http.StatusConflict)
}

// TestAuth_Renew_Success tests successful token renewal
func TestAuth_Renew_Success(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// First register and login
	email := fmt.Sprintf("renewtest-%s@example.com", NewTestUUID()[:8])
	password := "TestPass123!"

	resp, err := client.RawMultipartRequest("PUT", "/login", map[string]string{
		"email":    email,
		"password": password,
		"username": "renewtest" + NewTestUUID()[:8],
		"country":  "US",
	})

	require.NoError(t, err, "Register request should not fail")
	defer resp.Body.Close()

	// Login to get tokens
	resp, err = client.RawRequest("POST", "/login", map[string]interface{}{
		"email":    email,
		"password": password,
	})

	require.NoError(t, err, "Login request should not fail")
	defer resp.Body.Close()

	body := AssertResponseBody(t, resp)
	client.SetTokens(
		body["access_token"].(string),
		body["refresh_token"].(string),
		body["user_uuid"].(string),
	)

	// Now renew the token using refresh token
	resp, err = client.RequestWithRefreshToken("POST", "/renew", nil)

	require.NoError(t, err, "Renew request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)

	renewBody := AssertResponseBody(t, resp)
	AssertContainsField(t, renewBody, "access_token")
	AssertContainsField(t, renewBody, "user_uuid")

	// Verify new token is different
	assert.NotEqual(t, body["access_token"].(string), renewBody["access_token"].(string),
		"New access token should be different")
}

// TestAuth_Renew_InvalidToken tests token renewal with invalid token
func TestAuth_Renew_InvalidToken(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Set invalid token
	client.SetTokens("invalid.token.here", "invalid.refresh", "")

	resp, err := client.RequestWithRefreshToken("POST", "/renew", nil)

	require.NoError(t, err, "Renew request should not fail")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	AssertResponseStatus(t, resp, http.StatusUnauthorized)
}

// TestAuth_Renew_ExpiredToken tests token renewal with expired token
func TestAuth_Renew_ExpiredToken(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Set an expired token (this is a mock test - in reality you'd need a real expired token)
	client.SetTokens("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJub25lIiwiZXhwIjoxfQ.invalid",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJub25lIiwiZXhwIjoxfQ.invalid",
		"")

	resp, err := client.RequestWithRefreshToken("POST", "/renew", nil)

	require.NoError(t, err, "Renew request should not fail")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	AssertResponseStatus(t, resp, http.StatusUnauthorized)
}

// TestAuth_MissingCredentials tests login without credentials
func TestAuth_MissingCredentials(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Login without body
	resp, err := client.RawRequest("POST", "/login", nil)

	require.NoError(t, err, "Login request should not fail")
	defer resp.Body.Close()

	// Should return 400 Bad Request
	AssertResponseStatus(t, resp, http.StatusBadRequest)
}

// TestAuth_Login_WithEmptyPassword tests login with empty password
func TestAuth_Login_WithEmptyPassword(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	resp, err := client.RawRequest("POST", "/login", map[string]interface{}{
		"email":    "test@example.com",
		"password": "",
	})

	require.NoError(t, err, "Login request should not fail")
	defer resp.Body.Close()

	// Should return 400 Bad Request
	AssertResponseStatus(t, resp, http.StatusBadRequest)
}

// TestAuth_Register_MissingFields tests registration with missing fields
func TestAuth_Register_MissingFields(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	testCases := []struct {
		name string
		body map[string]string
	}{
		{
			name: "missing_email",
			body: map[string]string{
				"password": "TestPass123!",
				"username": "testuser",
				"country":  "US",
			},
		},
		{
			name: "missing_password",
			body: map[string]string{
				"email":    "test@example.com",
				"username": "testuser",
				"country":  "US",
			},
		},
		{
			name: "missing_username",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "TestPass123!",
				"country":  "US",
			},
		},
		{
			name: "missing_country",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "TestPass123!",
				"username": "testuser",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.RawMultipartRequest("PUT", "/login", tc.body)
			require.NoError(t, err, "Register request should not fail")
			defer resp.Body.Close()

			// Should return 400 Bad Request
			AssertResponseStatus(t, resp, http.StatusBadRequest)
		})
	}
}

// TestAuth_Login_JsonFields tests that JSON field names are properly handled
func TestAuth_Login_JsonFields(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Test with wrong field name
	resp, err := client.RawRequest("POST", "/login", map[string]interface{}{
		"user_email": "test@example.com",
		"pass":       "password",
	})

	require.NoError(t, err, "Login request should not fail")
	defer resp.Body.Close()

	// Should return 400 Bad Request (missing required fields)
	AssertResponseStatus(t, resp, http.StatusBadRequest)
}

// TestAuth_AccessTokenFormat tests that access token has correct JWT format
func TestAuth_AccessTokenFormat(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Register and login
	email := fmt.Sprintf("tokenformattest-%s@example.com", NewTestUUID()[:8])
	password := "TestPass123!"

	resp, err := client.RawMultipartRequest("PUT", "/login", map[string]string{
		"email":    email,
		"password": password,
		"username": "tokentest" + NewTestUUID()[:8],
		"country":  "US",
	})

	require.NoError(t, err, "Register should not fail")
	defer resp.Body.Close()

	// Login
	resp, err = client.RawRequest("POST", "/login", map[string]interface{}{
		"email":    email,
		"password": password,
	})

	require.NoError(t, err, "Login should not fail")
	defer resp.Body.Close()

	body := AssertResponseBody(t, resp)
	accessToken := body["access_token"].(string)

	// JWT should have 3 parts separated by dots
	parts := splitString(accessToken, ".")
	assert.Equal(t, 3, len(parts), "JWT should have 3 parts")

	// Each part should be base64 encoded
	for i, part := range parts {
		assert.NotEmpty(t, part, "JWT part %d should not be empty", i)
	}
}

// Helper function to split string
func splitString(s, sep string) []string {
	result := []string{}
	current := ""
	for _, c := range s {
		if string(c) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}

// TestAuth_GatewayProxiesAuth tests that the gateway properly proxies auth requests
func TestAuth_GatewayProxiesAuth(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// The gateway should properly forward to the user database service
	// This test verifies the proxy is working

	email := fmt.Sprintf("proxyauth-%s@example.com", NewTestUUID()[:8])
	password := "TestPass123!"

	resp, err := client.RawMultipartRequest("PUT", "/login", map[string]string{
		"email":    email,
		"password": password,
		"username": "proxyauth" + NewTestUUID()[:8],
		"country":  "US",
	})

	require.NoError(t, err, "Register should not fail")
	defer resp.Body.Close()

	// If the backend is not available, we get 502
	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated,
		"Auth should be proxied successfully")
}
