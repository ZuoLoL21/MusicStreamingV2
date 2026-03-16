//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"popularity/internal/di"
	"popularity/internal/handlers"

	libsdi "libs/di"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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
	VaultAddr       string
	VaultToken      string
	JWTStorePath    string
	ApplicationName string

	// Service
	ServicePort string
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
		JWTStorePath:       getEnv("JWT_STORE_PATH", "/tmp/jwt_stores"),
		ApplicationName:    getEnv("VAULT_APPLICATION_NAME", "popularity-system"),
		ServicePort:        getEnv("POPULARITY_PORT", "11090"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupClickHouse connects to ClickHouse and returns a connection
// It also cleans up all test data before and after the test
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

	// Clean database before test
	CleanupTestData(t, conn)

	t.Cleanup(func() {
		// Clean database after test
		CleanupTestData(t, conn)
		conn.Close()
	})

	return conn
}

// CleanupTestData removes all data from test tables
func CleanupTestData(t *testing.T, conn driver.Conn) {
	t.Helper()

	// Truncate materialized view intermediate tables first (they depend on base tables)
	materializedViews := []string{
		"track_popularity_inter",
		"artist_popularity_inter",
		"theme_popularity_inter",
		"track_by_theme_popularity_inter",
		"track_popularity_daily",
		"artist_popularity_daily",
		"theme_popularity_daily",
		"track_by_theme_popularity_daily",
	}

	for _, table := range materializedViews {
		query := fmt.Sprintf("TRUNCATE TABLE IF EXISTS %s", table)
		err := conn.Exec(context.Background(), query)
		if err != nil {
			t.Logf("Warning: Failed to truncate materialized view %s: %v", table, err)
		}
	}

	// Then truncate base tables
	baseTables := []string{
		"music_listen_events",
		"music_theme",
	}

	for _, table := range baseTables {
		query := fmt.Sprintf("TRUNCATE TABLE IF EXISTS %s", table)
		err := conn.Exec(context.Background(), query)
		if err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}

	// Wait for all truncations to complete
	time.Sleep(300 * time.Millisecond)
}

// SetupTestDatabase creates the required tables and materialized views for testing
func SetupTestDatabase(t *testing.T, conn driver.Conn) {
	t.Helper()

	// Create base tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS music_listen_events (
			event_time DateTime64(3, 'UTC') DEFAULT now64(3),
			user_uuid UUID,
			music_uuid UUID,
			artist_uuid UUID,
			album_uuid Nullable(UUID),
			listen_duration_seconds UInt32,
			track_duration_seconds UInt32,
			completion_ratio Float32
		) ENGINE = MergeTree() ORDER BY (user_uuid, event_time)`,

		`CREATE TABLE IF NOT EXISTS music_theme (
			music_uuid UUID,
			theme LowCardinality(String),
			views UInt64 DEFAULT 0,
			successes UInt64 DEFAULT 0,
			last_update DateTime64(3, 'UTC') DEFAULT now64(3)
		) ENGINE = ReplacingMergeTree() ORDER BY (music_uuid, theme)`,

		// Create daily aggregation tables
		`CREATE MATERIALIZED VIEW IF NOT EXISTS track_popularity_daily
		ENGINE = SummingMergeTree()
		ORDER BY (music_uuid, for_day)
		AS SELECT
			music_uuid,
			toDate(event_time) AS for_day,
			count() AS plays,
			sum(listen_duration_seconds) AS listen_seconds
		FROM music_listen_events
		GROUP BY music_uuid, for_day`,

		`CREATE MATERIALIZED VIEW IF NOT EXISTS artist_popularity_daily
		ENGINE = SummingMergeTree()
		ORDER BY (artist_uuid, for_day)
		AS SELECT
			artist_uuid,
			toDate(event_time) AS for_day,
			count() AS plays,
			sum(listen_duration_seconds) AS listen_seconds
		FROM music_listen_events
		GROUP BY artist_uuid, for_day`,

		`CREATE MATERIALIZED VIEW IF NOT EXISTS theme_popularity_daily
		ENGINE = SummingMergeTree()
		ORDER BY (theme, for_day)
		AS SELECT
			mt.theme,
			toDate(mle.event_time) AS for_day,
			count() AS plays,
			sum(mle.listen_duration_seconds) AS listen_seconds
		FROM music_listen_events mle
		INNER JOIN music_theme mt ON mle.music_uuid = mt.music_uuid
		GROUP BY mt.theme, for_day`,

		`CREATE MATERIALIZED VIEW IF NOT EXISTS track_by_theme_popularity_daily
		ENGINE = SummingMergeTree()
		ORDER BY (music_uuid, theme, for_day)
		AS SELECT
			mle.music_uuid,
			mt.theme,
			toDate(mle.event_time) AS for_day,
			count() AS plays,
			sum(mle.listen_duration_seconds) AS listen_seconds
		FROM music_listen_events mle
		INNER JOIN music_theme mt ON mle.music_uuid = mt.music_uuid
		GROUP BY mle.music_uuid, mt.theme, for_day`,

		// Create all-time views
		`CREATE MATERIALIZED VIEW IF NOT EXISTS track_popularity_inter
		ENGINE = SummingMergeTree()
		ORDER BY (music_uuid)
		AS SELECT
			music_uuid,
			sumState(1) AS decay_plays,
			sumState(listen_duration_seconds) AS decay_listen_seconds
		FROM music_listen_events
		GROUP BY music_uuid`,

		`CREATE MATERIALIZED VIEW IF NOT EXISTS artist_popularity_inter
		ENGINE = SummingMergeTree()
		ORDER BY (artist_uuid)
		AS SELECT
			artist_uuid,
			sumState(1) AS decay_plays,
			sumState(listen_duration_seconds) AS decay_listen_seconds
		FROM music_listen_events
		GROUP BY artist_uuid`,

		`CREATE MATERIALIZED VIEW IF NOT EXISTS theme_popularity_inter
		ENGINE = SummingMergeTree()
		ORDER BY (theme)
		AS SELECT
			mt.theme,
			sumState(1) AS decay_plays,
			sumState(mle.listen_duration_seconds) AS decay_listen_seconds
		FROM music_listen_events mle
		INNER JOIN music_theme mt ON mle.music_uuid = mt.music_uuid
		GROUP BY mt.theme`,

		`CREATE MATERIALIZED VIEW IF NOT EXISTS track_by_theme_popularity_inter
		ENGINE = SummingMergeTree()
		ORDER BY (music_uuid, theme)
		AS SELECT
			mle.music_uuid,
			mt.theme,
			sumState(1) AS decay_plays,
			sumState(mle.listen_duration_seconds) AS decay_listen_seconds
		FROM music_listen_events mle
		INNER JOIN music_theme mt ON mle.music_uuid = mt.music_uuid
		GROUP BY mle.music_uuid, mt.theme`,

		// Create final views
		`CREATE VIEW IF NOT EXISTS track_popularity AS
		SELECT
			music_uuid,
			sumMerge(decay_plays) AS decay_plays,
			sumMerge(decay_listen_seconds) AS decay_listen_seconds
		FROM track_popularity_inter
		GROUP BY music_uuid`,

		`CREATE VIEW IF NOT EXISTS artist_popularity AS
		SELECT
			artist_uuid,
			sumMerge(decay_plays) AS decay_plays,
			sumMerge(decay_listen_seconds) AS decay_listen_seconds
		FROM artist_popularity_inter
		GROUP BY artist_uuid`,

		`CREATE VIEW IF NOT EXISTS theme_popularity AS
		SELECT
			theme,
			sumMerge(decay_plays) AS decay_plays,
			sumMerge(decay_listen_seconds) AS decay_listen_seconds
		FROM theme_popularity_inter
		GROUP BY theme`,

		`CREATE VIEW IF NOT EXISTS track_by_theme_popularity AS
		SELECT
			music_uuid,
			theme,
			sumMerge(decay_plays) AS decay_plays,
			sumMerge(decay_listen_seconds) AS decay_listen_seconds
		FROM track_by_theme_popularity_inter
		GROUP BY music_uuid, theme`,
	}

	for _, q := range queries {
		err := conn.AsyncInsert(context.Background(), q, false)
		if err != nil {
			// Try synchronous execution for views
			err = conn.Exec(context.Background(), q)
		}
		if err != nil {
			t.Logf("Warning: Query may have failed: %v", err)
		}
	}

	// Wait for materialized views to be ready
	time.Sleep(2 * time.Second)
}

// CreateTestPopularityHandler creates a PopularityHandler for testing
func CreateTestPopularityHandler(t *testing.T) *handlers.PopularityHandler {
	t.Helper()
	logger := zap.NewNop()
	config := GetTestConfig()

	warehouseURL := fmt.Sprintf(
		"clickhouse://%s:%s@%s:%s/%s",
		config.ClickHouseUser,
		config.ClickHousePassword,
		config.ClickHouseHost,
		config.ClickHousePort,
		config.ClickHouseDB,
	)

	testConfig := &di.Config{
		WarehouseURL:    warehouseURL,
		ApplicationName: config.ApplicationName,
		JWTTimeout:      30 * time.Second,
		VaultAddr:       config.VaultAddr,
		VaultToken:      config.VaultToken,
	}

	returns := libsdi.NewReturnManager(logger)

	return handlers.NewPopularityHandler(logger, testConfig, returns)
}

// InsertTestListenEvent inserts a test listen event into ClickHouse
func InsertTestListenEvent(t *testing.T, conn driver.Conn, userUUID, musicUUID, artistUUID uuid.UUID, listenDuration uint32) {
	t.Helper()

	query := `
		INSERT INTO music_listen_events (user_uuid, music_uuid, artist_uuid, listen_duration_seconds, track_duration_seconds, completion_ratio)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	err := conn.Exec(context.Background(), query, userUUID, musicUUID, artistUUID, listenDuration, uint32(180), float32(listenDuration)/180.0)
	require.NoError(t, err, "Failed to insert listen event")
}

// WaitForMaterializedViews waits for materialized views to process data
// This is more reliable than a fixed sleep time
func WaitForMaterializedViews(t *testing.T, conn driver.Conn, expectedMinCount uint64) {
	t.Helper()

	maxAttempts := 20
	sleepDuration := 100 * time.Millisecond

	for i := 0; i < maxAttempts; i++ {
		var count uint64
		err := conn.QueryRow(context.Background(), "SELECT count() FROM track_popularity_inter").Scan(&count)
		if err == nil && count >= expectedMinCount {
			// Give a bit more time for all views to be consistent
			time.Sleep(100 * time.Millisecond)
			return
		}
		time.Sleep(sleepDuration)
	}

	t.Logf("Warning: Materialized views may not have fully updated after %v", time.Duration(maxAttempts)*sleepDuration)
}

// InsertTestTheme inserts a test music theme into ClickHouse
func InsertTestTheme(t *testing.T, conn driver.Conn, musicUUID uuid.UUID, theme string) {
	t.Helper()

	query := `
		INSERT INTO music_theme (music_uuid, theme)
		VALUES (?, ?)
	`

	err := conn.Exec(context.Background(), query, musicUUID, theme)
	require.NoError(t, err, "Failed to insert theme")
}

// NewTestRouter creates a mux.Router with the popularity handler routes
func NewTestRouter(handler *handlers.PopularityHandler) *mux.Router {
	r := mux.NewRouter()

	// All-time popularity endpoints
	r.HandleFunc("/popular/songs/all-time", handler.PopularSongsAllTime).Methods("GET")
	r.HandleFunc("/popular/artists/all-time", handler.PopularArtistAllTime).Methods("GET")
	r.HandleFunc("/popular/themes/all-time", handler.PopularThemeAllTime).Methods("GET")
	r.HandleFunc("/popular/songs/theme/{theme}", handler.PopularSongsAllTimeByTheme).Methods("GET")

	// Timeframe popularity endpoints
	r.HandleFunc("/popular/songs/timeframe", handler.PopularSongsTimeframe).Methods("GET")
	r.HandleFunc("/popular/artists/timeframe", handler.PopularArtistTimeframe).Methods("GET")
	r.HandleFunc("/popular/themes/timeframe", handler.PopularThemeTimeframe).Methods("GET")
	r.HandleFunc("/popular/songs/theme/{theme}/timeframe", handler.PopularSongsTimeframeByTheme).Methods("GET")

	return r
}

// NewTestUUID generates a new test UUID string
func NewTestUUID() string {
	return uuid.New().String()
}

// GetClickHouseDB returns a *sql.DB for direct database access
func GetClickHouseDB(t *testing.T) *sql.DB {
	t.Helper()
	config := GetTestConfig()

	warehouseURL := fmt.Sprintf(
		"clickhouse://%s:%s@%s:%s/%s",
		config.ClickHouseUser,
		config.ClickHousePassword,
		config.ClickHouseHost,
		config.ClickHousePort,
		config.ClickHouseDB,
	)

	db, err := sql.Open("clickhouse", warehouseURL)
	require.NoError(t, err, "Failed to connect to ClickHouse")

	err = db.Ping()
	require.NoError(t, err, "Failed to ping ClickHouse")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}
