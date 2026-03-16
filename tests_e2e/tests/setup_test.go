//go:build e2e

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig holds configuration for E2E tests
type TestConfig struct {
	// Gateway API
	GatewayBaseURL string

	// Service URLs (for direct testing)
	EventIngestionURL string
	PopularityURL     string
	RecommendationURL string

	// External Services
	VaultAddr  string
	VaultToken string

	// Test credentials
	TestUserEmail    string
	TestUserPassword string
}

// GetTestConfig returns test configuration from environment variables
func GetTestConfig() *TestConfig {
	return &TestConfig{
		GatewayBaseURL:    getEnv("GATEWAY_URL", "http://localhost:8080"),
		EventIngestionURL: getEnv("EVENT_INGESTION_URL", "http://localhost:11080"),
		PopularityURL:     getEnv("POPULARITY_URL", "http://localhost:11090"),
		RecommendationURL: getEnv("RECOMMENDATION_URL", "http://localhost:11070"),
		VaultAddr:         getEnv("VAULT_ADDR", "http://localhost:18200"),
		VaultToken:        getEnv("VAULT_TOKEN", "test-root-token"),
		TestUserEmail:     getEnv("TEST_USER_EMAIL", "test@example.com"),
		TestUserPassword:  getEnv("TEST_USER_PASSWORD", "testpassword123"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestClient holds authenticated HTTP client state
type TestClient struct {
	baseURL       string
	httpClient    *http.Client
	accessToken   string
	refreshToken  string
	userUUID      string
	mu            sync.Mutex
	serviceTokens map[string]string // For internal service JWTs
}

// NewTestClient creates a new test client
func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		baseURL:       baseURL,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		serviceTokens: make(map[string]string),
	}
}

// SetTokens sets authentication tokens
func (c *TestClient) SetTokens(accessToken, refreshToken, userUUID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = accessToken
	c.refreshToken = refreshToken
	c.userUUID = userUUID
}

// GetAccessToken returns the current access token
func (c *TestClient) GetAccessToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.accessToken
}

// GetUserUUID returns the current user UUID
func (c *TestClient) GetUserUUID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.userUUID
}

// Request makes an authenticated HTTP request
func (c *TestClient) Request(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add auth header if we have a token
	if token := c.GetAccessToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return c.httpClient.Do(req)
}

// RequestWithServiceToken makes a request with service JWT
func (c *TestClient) RequestWithServiceToken(serviceName, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add service token if available
	c.mu.Lock()
	if token, ok := c.serviceTokens[serviceName]; ok {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	c.mu.Unlock()

	return c.httpClient.Do(req)
}

// RawRequest makes a raw HTTP request without auth
func (c *TestClient) RawRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// Response represents a parsed API response
type Response struct {
	StatusCode int
	Body       map[string]interface{}
	RawBody    string
}

// ParseResponse parses an HTTP response into a Response struct
func ParseResponse(resp *http.Response) (*Response, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &Response{
		StatusCode: resp.StatusCode,
		RawBody:    string(body),
	}

	// Try to parse JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		result.Body = parsed
	}

	return result, nil
}

// Helper functions for common test operations

// NewTestUUID generates a new test UUID string
func NewTestUUID() string {
	return uuid.New().String()
}

// CreateTestUser creates a test user and returns the credentials
func CreateTestUser(t *testing.T, client *TestClient, prefix string) (email, password string) {
	email = fmt.Sprintf("%s-%s@example.com", prefix, NewTestUUID()[:8])
	password = "TestPass123!"

	// Register the user
	resp, err := client.RawRequest("PUT", "/login", map[string]interface{}{
		"email":        email,
		"password":     password,
		"username":     prefix + NewTestUUID()[:8],
		"display_name": "Test User",
	})

	require.NoError(t, err, "Failed to create test user request")
	defer resp.Body.Close()

	// Accept 200 or 201
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		t.Logf("User might already exist, response: %d", resp.StatusCode)
	}

	return email, password
}

// LoginUser logs in a user and returns the tokens
func LoginUser(t *testing.T, client *TestClient, email, password string) (accessToken, refreshToken, userUUID string) {
	resp, err := client.RawRequest("POST", "/login", map[string]interface{}{
		"email":    email,
		"password": password,
	})

	require.NoError(t, err, "Failed to login")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Login should succeed")

	response, err := ParseResponse(resp)
	require.NoError(t, err, "Failed to parse login response")
	require.NotNil(t, response.Body, "Response body should not be nil")

	accessToken = response.Body["access_token"].(string)
	refreshToken = response.Body["refresh_token"].(string)
	userUUID = response.Body["user_uuid"].(string)

	client.SetTokens(accessToken, refreshToken, userUUID)

	return accessToken, refreshToken, userUUID
}

// RefreshToken refreshes the access token
func RefreshToken(t *testing.T, client *TestClient) string {
	client.mu.Lock()
	client.mu.Unlock()

	resp, err := client.RawRequest("POST", "/renew", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Token refresh should succeed")

	response, err := ParseResponse(resp)
	require.NoError(t, err)

	accessToken := response.Body["access_token"].(string)
	client.mu.Lock()
	client.accessToken = accessToken
	client.mu.Unlock()

	return accessToken
}

// CleanupUser cleans up a test user
func CleanupUser(t *testing.T, client *TestClient, userUUID string) {
	// Note: In a real implementation, you would delete the user
	// For now, we just clear the tokens
	client.SetTokens("", "", "")
}

// AssertResponseOK asserts that the response is 200 OK
func AssertResponseOK(t *testing.T, resp *http.Response) {
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Response should be 200 OK")
}

// AssertResponseStatus asserts the response has the expected status
func AssertResponseStatus(t *testing.T, resp *http.Response, expectedStatus int) {
	assert.Equal(t, expectedStatus, resp.StatusCode, fmt.Sprintf("Response should be %d", expectedStatus))
}

// AssertResponseBody parses and returns the response body
func AssertResponseBody(t *testing.T, resp *http.Response) map[string]interface{} {
	response, err := ParseResponse(resp)
	require.NoError(t, err)
	require.NotNil(t, response.Body, "Response body should not be nil")
	return response.Body
}

// AssertContainsField asserts that the response contains the specified field
func AssertContainsField(t *testing.T, body map[string]interface{}, fieldName string) {
	_, exists := body[fieldName]
	assert.True(t, exists, fmt.Sprintf("Response should contain field: %s", fieldName))
}

// AssertNoError checks for error field in response
func AssertNoError(t *testing.T, body map[string]interface{}) {
	if err, exists := body["error"]; exists {
		assert.Fail(t, fmt.Sprintf("Response contains error: %v", err))
	}
}

// TestMainSetup sets up the test environment
func TestMainSetup(t *testing.T) {
	config := GetTestConfig()

	// Verify connectivity
	client := NewTestClient(config.GatewayBaseURL)
	resp, err := client.RawRequest("GET", "/health", nil)
	if err != nil {
		t.Skipf("Gateway API not available: %v", err)
	}
	resp.Body.Close()
}

// SetupAuthenticatedClient sets up an authenticated test client
func SetupAuthenticatedClient(t *testing.T, config *TestConfig) *TestClient {
	client := NewTestClient(config.GatewayBaseURL)

	// Try to login with test user
	email := config.TestUserEmail
	password := config.TestUserPassword

	resp, err := client.RawRequest("POST", "/login", map[string]interface{}{
		"email":    email,
		"password": password,
	})

	if err != nil || resp.StatusCode != http.StatusOK {
		// Try to register
		resp, err = client.RawRequest("PUT", "/login", map[string]interface{}{
			"email":        email,
			"password":     password,
			"username":     "testuser",
			"display_name": "Test User",
		})
		if err != nil {
			t.Skipf("Cannot authenticate: %v", err)
		}
		resp.Body.Close()

		// Try login again
		resp, err = client.RawRequest("POST", "/login", map[string]interface{}{
			"email":    email,
			"password": password,
		})
		if err != nil {
			t.Skipf("Cannot authenticate: %v", err)
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skipf("Cannot authenticate: status %d", resp.StatusCode)
	}

	response, err := ParseResponse(resp)
	require.NoError(t, err)

	client.SetTokens(
		response.Body["access_token"].(string),
		response.Body["refresh_token"].(string),
		response.Body["user_uuid"].(string),
	)

	return client
}

// WaitForService waits for a service to be available
func WaitForService(url string, timeout time.Duration) error {
	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()

	for time.Since(start) < timeout {
		resp, err := client.Get(url + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("service %s not available after %v", url, timeout)
}

// GenerateTestMusicData generates test music metadata
func GenerateTestMusicData() map[string]interface{} {
	return map[string]interface{}{
		"title":       "Test Song " + NewTestUUID()[:8],
		"description": "Test description",
		"duration":    180,
		"genre":       "pop",
		"tags":        []string{"test", "music"},
	}
}

// GenerateTestAlbumData generates test album metadata
func GenerateTestAlbumData() map[string]interface{} {
	return map[string]interface{}{
		"name":         "Test Album " + NewTestUUID()[:8],
		"description":  "Test album description",
		"release_date": "2024-01-01",
		"genre":        "pop",
	}
}

// GenerateTestPlaylistData generates test playlist metadata
func GenerateTestPlaylistData() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Test Playlist " + NewTestUUID()[:8],
		"description": "Test playlist description",
		"is_public":   false,
	}
}

// GenerateTestArtistData generates test artist metadata
func GenerateTestArtistData() map[string]interface{} {
	return map[string]interface{}{
		"name":  "Test Artist " + NewTestUUID()[:8],
		"bio":   "Test artist bio",
		"genre": "pop",
	}
}

// Helper to extract UUID from response
func ExtractUUID(t *testing.T, body map[string]interface{}, field string) string {
	value, ok := body[field]
	require.True(t, ok, fmt.Sprintf("Field %s should exist", field))
	return value.(string)
}

// Helper to extract string from response
func ExtractString(t *testing.T, body map[string]interface{}, field string) string {
	value, ok := body[field]
	require.True(t, ok, fmt.Sprintf("Field %s should exist", field))
	str, ok := value.(string)
	require.True(t, ok, fmt.Sprintf("Field %s should be a string", field))
	return str
}

// Helper to extract bool from response
func ExtractBool(t *testing.T, body map[string]interface{}, field string) bool {
	value, ok := body[field]
	require.True(t, ok, fmt.Sprintf("Field %s should exist", field))
	b, ok := value.(bool)
	require.True(t, ok, fmt.Sprintf("Field %s should be a bool", field))
	return b
}

// Helper to extract array from response
func ExtractArray(t *testing.T, body map[string]interface{}, field string) []interface{} {
	value, ok := body[field]
	require.True(t, ok, fmt.Sprintf("Field %s should exist", field))
	arr, ok := value.([]interface{})
	require.True(t, ok, fmt.Sprintf("Field %s should be an array", field))
	return arr
}

// CreateAuthenticatedRouter creates a test router with auth middleware
func CreateAuthenticatedRouter(handler http.Handler, token string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+token)
		handler.ServeHTTP(w, r)
	}))
}

// Context with timeout
func ContextWithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}
