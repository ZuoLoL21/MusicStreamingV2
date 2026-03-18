//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"event_ingestion/internal/di"
	"event_ingestion/internal/handlers"
	"fmt"
	"io"
	libsdi "libs/di"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	// ClickHouse
	ClickHouseHost     string
	ClickHousePort     string
	ClickHouseUser     string
	ClickHousePassword string
	ClickHouseDB       string

	// Vault
	VaultAddr  string
	VaultToken string

	// Service
	ServiceURL string
}

// GetTestConfig returns test configuration from environment variables
func GetTestConfig() *TestConfig {
	return &TestConfig{
		ClickHouseHost:     getEnv("CLICKHOUSE_HOST", "localhost"),
		ClickHousePort:     getEnv("CLICKHOUSE_PORT", "9000"),
		ClickHouseUser:     getEnv("CLICKHOUSE_USER", "clickhouse"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", "clickhouse"),
		ClickHouseDB:       getEnv("CLICKHOUSE_DB", "default"),
		VaultAddr:          getEnv("VAULT_ADDR", "http://localhost:18200"),
		VaultToken:         getEnv("VAULT_TOKEN", "test-root-token"),
		ServiceURL:         getEnv("SERVICE_URL", "http://localhost:11080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupClickHouse connects to ClickHouse and returns a connection
func SetupClickHouse(t *testing.T) driver.Conn {
	t.Helper()
	config := GetTestConfig()

	connString := fmt.Sprintf(
		"clickhouse://%s:%s@%s:%s/%s",
		config.ClickHouseUser,
		config.ClickHousePassword,
		config.ClickHouseHost,
		config.ClickHousePort,
		config.ClickHouseDB,
	)

	opts, err := clickhouse.ParseDSN(connString)
	require.NoError(t, err, "failed to parse ClickHouse DSN")

	conn, err := clickhouse.Open(opts)
	require.NoError(t, err, "failed to connect to ClickHouse")

	// Test connection
	err = conn.Ping(context.Background())
	require.NoError(t, err, "failed to ping ClickHouse")

	t.Cleanup(func() {
		conn.Close()
	})

	return conn
}

// NewTestClickHouseClient creates a ClickHouse client for testing
func NewTestClickHouseClient(t *testing.T) *di.ClickHouseClient {
	t.Helper()
	config := GetTestConfig()
	logger := zap.NewNop()

	connString := fmt.Sprintf(
		"clickhouse://%s:%s@%s:%s/%s",
		config.ClickHouseUser,
		config.ClickHousePassword,
		config.ClickHouseHost,
		config.ClickHousePort,
		config.ClickHouseDB,
	)

	testConfig := &di.Config{
		ClickHouseConnectionString: connString,
	}

	client, err := di.NewClickHouseClient(testConfig, logger)
	require.NoError(t, err, "failed to create ClickHouse test client")
	require.NotNil(t, client, "ClickHouse client should not be nil")

	return client
}

// SetupVaultClient sets up and returns Vault configuration for testing
func SetupVaultClient(t *testing.T) (vaultAddr, vaultToken string) {
	t.Helper()
	config := GetTestConfig()
	return config.VaultAddr, config.VaultToken
}

// testVaultConfig implements the vault config interface for testing
type testVaultConfig struct {
	vaultAddr  string
	vaultToken string
}

func (c *testVaultConfig) GetVaultAddr() string        { return c.vaultAddr }
func (c *testVaultConfig) GetVaultToken() string       { return c.vaultToken }
func (c *testVaultConfig) GetVaultTransitPath() string { return "transit" }

// NewTestVaultConfig creates a test vault configuration
func NewTestVaultConfig(t *testing.T) *testVaultConfig {
	t.Helper()
	vaultAddr, vaultToken := SetupVaultClient(t)
	return &testVaultConfig{
		vaultAddr:  vaultAddr,
		vaultToken: vaultToken,
	}
}

// CreateTestHandler creates an EventHandler for testing
func CreateTestHandler(t *testing.T) *handlers.EventHandler {
	t.Helper()
	logger := zap.NewNop()
	config := GetTestConfig()

	connString := fmt.Sprintf(
		"clickhouse://%s:%s@%s:%s/%s",
		config.ClickHouseUser,
		config.ClickHousePassword,
		config.ClickHouseHost,
		config.ClickHousePort,
		config.ClickHouseDB,
	)

	testConfig := &di.Config{
		ClickHouseConnectionString: connString,
	}

	clickhouse, err := di.NewClickHouseClient(testConfig, logger)
	require.NoError(t, err, "failed to create ClickHouse client for handler")

	returns := libsdi.NewReturnManager(logger)

	return handlers.NewEventHandler(testConfig, returns, clickhouse)
}

// createJSONRequest creates an HTTP request with JSON body
func createJSONRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// createRequest creates an HTTP request for testing
func createRequest(t *testing.T, method, path string, body io.Reader) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// assertJSONResponse validates HTTP response status and parses JSON body
func assertJSONResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int, target interface{}) {
	t.Helper()
	assert.Equal(t, expectedStatus, rr.Code, "Response body: %s", rr.Body.String())
	if target != nil && rr.Code == expectedStatus {
		err := json.Unmarshal(rr.Body.Bytes(), target)
		require.NoError(t, err, "Failed to unmarshal response JSON")
	}
}

// assertErrorResponse validates error response with specific message pattern
func assertErrorResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int, messageContains string) {
	t.Helper()
	assert.Equal(t, expectedStatus, rr.Code)
	if messageContains != "" {
		assert.Contains(t, rr.Body.String(), messageContains)
	}
}

// createTestRouter creates a mux.Router with the event handler routes
func createTestRouter(handler *handlers.EventHandler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/events/listen", handler.IngestListenEvent).Methods("POST")
	r.HandleFunc("/events/like", handler.IngestLikeEvent).Methods("POST")
	r.HandleFunc("/events/theme", handler.IngestThemeEvent).Methods("POST")
	r.HandleFunc("/events/user", handler.IngestUserDimEvent).Methods("POST")
	return r
}

// EnsureTimestampDistinct adds minimal delay for timestamp uniqueness
func ensureTimestampDistinct() {
	time.Sleep(10 * time.Millisecond)
}

// NewTestUUID generates a new test UUID string
func NewTestUUID() string {
	return uuid.New().String()
}
