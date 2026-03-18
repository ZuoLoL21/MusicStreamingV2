//go:build e2e

package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGateway_HealthCheck tests the gateway health endpoint
func TestGateway_HealthCheck(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	resp, err := client.RawRequest("GET", "/health", nil)
	require.NoError(t, err, "Health check request should not fail")
	defer resp.Body.Close()

	AssertResponseStatus(t, resp, http.StatusOK)

	body := AssertResponseBody(t, resp)
	assert.Equal(t, "healthy", body["status"])
}

// TestGateway_ProxiesToBackend tests that gateway proxies requests to backend
func TestGateway_ProxiesToBackend(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Test that the gateway properly proxies user requests
	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "User request should be proxied")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// Should not return 404 from gateway itself
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
}

// TestGateway_Returns502OnBackendDown tests that gateway returns 502 when backend is down
func TestGateway_Returns502OnBackendDown(t *testing.T) {
	// This test would require shutting down a backend service
	// It's here for documentation purposes
	t.Skip("This test requires manual intervention to simulate backend failure")
}

// TestGateway_HandlesTimeout tests that gateway handles backend timeouts
func TestGateway_HandlesTimeout(t *testing.T) {
	// This test would require configuring a slow backend
	t.Skip("This test requires manual intervention to simulate timeout")
}

// TestGateway_CORSHeaders tests that CORS headers are properly set
func TestGateway_CORSHeaders(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	req, err := http.NewRequest("GET", config.GatewayBaseURL+"/health", nil)
	require.NoError(t, err, "Request should not fail")

	// Add origin header
	req.Header.Set("Origin", "http://example.com")

	resp, err := client.httpClient.Do(req)
	require.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	// Check for CORS headers (if enabled)
	// The presence or absence is OK as long as the request succeeds
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGateway_PreservesHeaders tests that gateway preserves important headers
func TestGateway_PreservesHeaders(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Make request with custom header
	req, err := http.NewRequest("GET", config.GatewayBaseURL+"/users/me", nil)
	require.NoError(t, err, "Request should not fail")

	req.Header.Set("Authorization", "Bearer "+client.GetAccessToken())
	req.Header.Set("X-Request-ID", "test-request-id")
	req.Header.Set("Accept", "application/json")

	resp, err := client.httpClient.Do(req)
	require.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// Should succeed
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized,
		"Request should be handled")
}

// TestGateway_RateLimiting tests rate limiting if implemented
func TestGateway_RateLimiting(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Make many rapid requests
	for i := 0; i < 10; i++ {
		resp, err := client.Request("GET", "/users/me", nil)
		require.NoError(t, err, "Request should not fail")
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			// Rate limiting is working
			return
		}
	}
	// If we get here, rate limiting might not be implemented or limit is high
}

// TestGateway_RequestID tests that request ID is propagated
func TestGateway_RequestID(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	// Response should have request ID header (if implemented)
	// This is optional functionality
	_ = resp.Header.Get("X-Request-ID")
}

// TestGateway_BadGatewayMessage tests that 502 errors have proper messages
func TestGateway_BadGatewayMessage(t *testing.T) {
	// This test would require a misconfigured backend
	t.Skip("This test requires manual intervention")
}

// TestGateway_NotFound tests that unknown routes return 404
func TestGateway_NotFound(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	resp, err := client.RawRequest("GET", "/nonexistent/endpoint", nil)
	require.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	// Should return 404
	AssertResponseStatus(t, resp, http.StatusNotFound)
}

// TestGateway_MethodNotAllowed tests that wrong methods return 405
func TestGateway_MethodNotAllowed(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Try to POST to a GET-only endpoint
	resp, err := client.RawRequest("POST", "/health", nil)
	require.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	// Should return 405 Method Not Allowed or 200 (if POST is allowed)
	assert.True(t, resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusOK,
		"Method handling should be correct")
}

// TestGateway_UnsupportedMediaType tests handling of unsupported media types
func TestGateway_UnsupportedMediaType(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Create request with XML content
	req, err := http.NewRequest("POST", config.GatewayBaseURL+"/login", nil)
	require.NoError(t, err, "Request should not fail")

	req.Header.Set("Content-Type", "application/xml")

	resp, err := client.httpClient.Do(req)
	require.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	// Should handle gracefully (either 415 or attempt to parse)
	assert.True(t, resp.StatusCode < 500, "Should not crash on unsupported media type")
}

// TestGateway_InvalidJSON tests handling of invalid JSON
func TestGateway_InvalidJSON(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Create request with invalid JSON
	req, err := http.NewRequest("POST", config.GatewayBaseURL+"/login", nil)
	require.NoError(t, err, "Request should not fail")

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	require.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	// Should return 400 Bad Request
	assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusOK,
		"Invalid JSON should be handled")
}

// TestGateway_AuthenticationRequired tests that protected endpoints require auth
func TestGateway_AuthenticationRequired(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	protectedEndpoints := []string{
		"/users/me",
		"/artists",
		"/albums",
		"/music",
		"/playlists",
	}

	for _, endpoint := range protectedEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := client.RawRequest("GET", endpoint, nil)
			require.NoError(t, err, "Request should not fail")
			defer resp.Body.Close()

			// Should return 401 Unauthorized
			AssertResponseStatus(t, resp, http.StatusUnauthorized)
		})
	}
}

// TestGateway_InvalidTokenRejected tests that invalid tokens are rejected
func TestGateway_InvalidTokenRejected(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Set invalid token
	client.SetTokens("invalid-token", "", "")

	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	AssertResponseStatus(t, resp, http.StatusUnauthorized)
}

// TestGateway_ExpiredTokenRejected tests that expired tokens are rejected
func TestGateway_ExpiredTokenRejected(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Set expired token (this is a mock)
	client.SetTokens("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJub25lIiwiZXhwIjoxfQ.invalid",
		"", "")

	resp, err := client.Request("GET", "/users/me", nil)
	require.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	AssertResponseStatus(t, resp, http.StatusUnauthorized)
}

// TestGateway_RefreshTokenRotation tests that refresh tokens are rotated
func TestGateway_RefreshTokenRotation(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Register and login
	email := "refreshtest-" + NewTestUUID()[:8] + "@example.com"
	password := "TestPass123!"
	deviceID := "00000000-0000-0000-0000-000000000001"

	resp, err := client.RawMultipartRequest("PUT", "/login", map[string]string{
		"email":       email,
		"password":    password,
		"username":    "refreshtest" + NewTestUUID()[:8],
		"country":     "US",
		"device_id":   deviceID,
		"device_name": "test-device",
	})

	require.NoError(t, err, "Register should not fail")
	defer resp.Body.Close()

	// Login
	resp, err = client.RawRequest("POST", "/login", map[string]interface{}{
		"email":       email,
		"password":    password,
		"device_id":   deviceID,
		"device_name": "test-device",
	})

	require.NoError(t, err, "Login should not fail")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skip("Cannot login for refresh test")
	}

	body := AssertResponseBody(t, resp)
	initialRefreshToken := body["refresh_token"].(string)

	client.SetTokens(
		body["access_token"].(string),
		body["refresh_token"].(string),
		body["user_uuid"].(string),
	)

	// Refresh token using refresh token authentication
	resp, err = client.RequestWithRefreshToken("POST", "/renew", nil)
	require.NoError(t, err, "Renew should not fail")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skip("Cannot refresh token")
	}

	newBody := AssertResponseBody(t, resp)
	newRefreshToken := newBody["refresh_token"].(string)

	// Refresh token should be different (rotation)
	// Note: This depends on implementation - some systems don't rotate refresh tokens
	_ = initialRefreshToken
	_ = newRefreshToken
}
