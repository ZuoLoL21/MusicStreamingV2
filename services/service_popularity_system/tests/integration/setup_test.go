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

	err := conn.AsyncInsert(context.Background(), query, false, userUUID, musicUUID, artistUUID, listenDuration, uint32(180), float32(listenDuration)/180.0)
	if err != nil {
		err = conn.Exec(context.Background(), query, userUUID, musicUUID, artistUUID, listenDuration, uint32(180), float32(listenDuration)/180.0)
	}
	require.NoError(t, err, "Failed to insert listen event")

	// Wait for the event to be processed
	time.Sleep(100 * time.Millisecond)
}

// InsertTestTheme inserts a test music theme into ClickHouse
func InsertTestTheme(t *testing.T, conn driver.Conn, musicUUID uuid.UUID, theme string) {
	t.Helper()

	query := `
		INSERT INTO music_theme (music_uuid, theme)
		VALUES (?, ?)
	`

	err := conn.AsyncInsert(context.Background(), query, false, musicUUID, theme)
	if err != nil {
		err = conn.Exec(context.Background(), query, musicUUID, theme)
	}
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
