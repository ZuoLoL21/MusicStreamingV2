//go:build integration

package integration

import (
	"backend/internal/client"
	sqlhandler "backend/sql/sqlc"
	"backend/tests/integration/builders"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	libsconsts "libs/consts"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// MinIO
	MinIOEndpoint       string
	MinIOPublicEndpoint string
	MinIOAccessKey      string
	MinIOSecretKey      string
	MinIOBucket         string
	MinIOUseSSL         bool

	// Vault
	VaultAddr  string
	VaultToken string

	// Mock services
	EventIngestionURL string
}

// GetTestConfig returns test configuration from environment variables
func GetTestConfig() *TestConfig {
	return &TestConfig{
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "15432"),
		DBUser:              getEnv("DB_USER", "test_user"),
		DBPassword:          getEnv("DB_PASSWORD", "test_password"),
		DBName:              getEnv("DB_NAME", "music_streaming_test"),
		DBSSLMode:           getEnv("DB_SSLMODE", "disable"),
		MinIOEndpoint:       getEnv("MINIO_ENDPOINT", "localhost:19000"),
		MinIOPublicEndpoint: getEnv("MINIO_PUBLIC_ENDPOINT", "localhost:19000"),
		MinIOAccessKey:      getEnv("MINIO_ACCESS_KEY", "testuser"),
		MinIOSecretKey:      getEnv("MINIO_SECRET_KEY", "testpassword"),
		MinIOBucket:         getEnv("MINIO_BUCKET_NAME", "test-bucket"),
		MinIOUseSSL:         getEnv("MINIO_USE_SSL", "false") == "true",
		VaultAddr:           getEnv("VAULT_ADDR", "http://localhost:18200"),
		VaultToken:          getEnv("VAULT_TOKEN", "test-root-token"),
		EventIngestionURL:   getEnv("EVENT_INGESTION_URL", "http://localhost:11080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupTestDB connects to the test database, runs migrations, and returns a connection pool
func SetupTestDB(t *testing.T) (*pgxpool.Pool, *sqlhandler.Queries) {
	t.Helper()
	config := GetTestConfig()

	// Run migrations
	runMigrations(t, config)

	// Connect to database
	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName, config.DBSSLMode,
	)

	pool, err := pgxpool.New(context.Background(), connString)
	require.NoError(t, err, "failed to connect to test database")

	// Test connection
	err = pool.Ping(context.Background())
	require.NoError(t, err, "failed to ping test database")

	// Set pg_trgm similarity threshold for search queries
	// Lower threshold (0.2 instead of default 0.3) allows shorter queries to match longer strings
	_, err = pool.Exec(context.Background(), "SELECT set_limit(0.2);")
	require.NoError(t, err, "failed to set pg_trgm similarity threshold")

	// Create queries instance
	queries := sqlhandler.New(pool)

	// Register cleanup
	t.Cleanup(func() {
		pool.Close()
	})

	return pool, queries
}

// runMigrations runs database migrations using psql
func runMigrations(t *testing.T, config *TestConfig) {
	t.Helper()

	// Schema directory path - go back from tests/integration to service root
	schemaDir := filepath.Join("..", "..", "sql", "schema")

	// Get all SQL files in order
	sqlFiles := []string{
		filepath.Join(schemaDir, "00_init.sql"),
		filepath.Join(schemaDir, "01_tables.sql"),
		filepath.Join(schemaDir, "02_triggers.sql"),
		filepath.Join(schemaDir, "03_functions.sql"),
		// Note: Skip 04_seed.sql for tests - we'll create our own test data
	}

	connString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName, config.DBSSLMode,
	)

	for _, sqlFile := range sqlFiles {
		if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
			t.Logf("Skipping migration file (not found): %s", sqlFile)
			continue
		}

		t.Logf("Running migration: %s", sqlFile)
		cmd := exec.Command("psql", connString, "-f", sqlFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Migration output: %s", string(output))
			require.NoError(t, err, "failed to run migration: %s", sqlFile)
		}
	}
}

// SetupMinIOClient creates and returns a MinIO client for testing
func SetupMinIOClient(t *testing.T) *client.MinIOFileStorageClient {
	t.Helper()
	config := GetTestConfig()
	logger := zap.NewNop()

	minioClient, err := client.NewMinIOFileStorageClient(
		config.MinIOEndpoint,
		config.MinIOAccessKey,
		config.MinIOSecretKey,
		config.MinIOBucket,
		logger,
	)
	require.NoError(t, err, "failed to create MinIO test client")
	require.NotNil(t, minioClient, "MinIO client should not be nil")

	return minioClient
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
func (c *testVaultConfig) GetJWTTimeout() time.Duration {
	return 5 * time.Second
}

// NewTestVaultConfig creates a test vault configuration
func NewTestVaultConfig(t *testing.T) *testVaultConfig {
	t.Helper()
	vaultAddr, vaultToken := SetupVaultClient(t)
	return &testVaultConfig{
		vaultAddr:  vaultAddr,
		vaultToken: vaultToken,
	}
}

// CleanupTestData removes all data from test tables
func CleanupTestData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	// Disable triggers temporarily to avoid foreign key issues
	_, err := pool.Exec(context.Background(), "SET session_replication_role = 'replica';")
	require.NoError(t, err)

	// Clean up tables in reverse dependency order
	tables := []string{
		"playlist_track",    // Depends on: playlist, music
		"playlist",          // Depends on: users
		"tag_assignment",    // Junction: music ↔ music_tags
		"music_tags",        // Tag definitions
		"listening_history", // Depends on: users, music
		"likes",             // Depends on: users, music
		"follows",           // Depends on: users, artist (single table for both user→user and user→artist)
		"music",             // Depends on: artist, album, users
		"album",             // Depends on: artist
		"artist_member",     // Depends on: artist, users
		"artist",            // Base table
		"users",             // Base table
	}

	for _, table := range tables {
		_, err := pool.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s CASCADE;", table))
		if err != nil {
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}

	// Re-enable triggers
	_, err = pool.Exec(context.Background(), "SET session_replication_role = 'origin';")
	require.NoError(t, err)
}

// wrapWithAuth wraps an HTTP handler with authentication context
// This simulates what the real service JWT middleware does
func wrapWithAuth(t *testing.T, handler http.HandlerFunc, userUUID pgtype.UUID) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		uuidStr := builders.UUIDToString(userUUID)
		ctx := context.WithValue(r.Context(), libsconsts.UserUUIDKey, uuidStr)
		handler(w, r.WithContext(ctx))
	}
}

// createRequest creates an HTTP request for testing
func createRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		if r, ok := body.(io.Reader); ok {
			bodyReader = r
		}
	}

	req := httptest.NewRequest(method, path, bodyReader)
	// Set Content-Type header for JSON requests
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// createJSONRequest creates an HTTP request with JSON body
func createJSONRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// ensureTimestampDistinct adds minimal delay for timestamp uniqueness
func ensureTimestampDistinct() {
	time.Sleep(10 * time.Millisecond)
}

// createTestImage creates a minimal valid PNG image for testing (10x10 pixels)
func createTestImage(width, height int) []byte {
	// Create a minimal PNG image
	buf := &bytes.Buffer{}
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a simple color (optional, makes it slightly larger but more realistic)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}

	_ = png.Encode(buf, img)
	return buf.Bytes()
}

// createMultipartRequest creates an HTTP request with multipart form data for file uploads
func createMultipartRequest(t *testing.T, method, path string, fileFieldName, fileName string, fileContent []byte, formFields map[string]string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file if provided
	if fileFieldName != "" && fileName != "" {
		part, err := writer.CreateFormFile(fileFieldName, fileName)
		require.NoError(t, err)
		_, err = part.Write(fileContent)
		require.NoError(t, err)
	}

	// Add form fields
	for key, val := range formFields {
		err := writer.WriteField(key, val)
		require.NoError(t, err)
	}

	err := writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
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
