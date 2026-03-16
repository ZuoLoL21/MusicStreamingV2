//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ListenEvent_Ingestion(t *testing.T) {
	conn := SetupClickHouseForTest(t)
	handler := CreateTestHandler(t)

	userUUID := NewTestUUID()
	musicUUID := NewTestUUID()
	artistUUID := NewTestUUID()
	albumUUID := NewTestUUID()

	// Insert listen event via handler
	reqBody := map[string]interface{}{
		"user_uuid":               userUUID,
		"music_uuid":              musicUUID,
		"artist_uuid":             artistUUID,
		"album_uuid":              albumUUID,
		"listen_duration_seconds": 120,
		"track_duration_seconds":  180,
		"completion_ratio":        0.67,
	}

	req := createJSONRequest(t, "POST", "/events/listen", reqBody)
	router := mux.NewRouter()
	router.HandleFunc("/events/listen", handler.IngestListenEvent).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["success"])

	// Verify data was inserted into ClickHouse
	var result struct {
		UserUUID              string
		MusicUUID             string
		ArtistUUID            string
		AlbumUUID             string
		ListenDurationSeconds uint32
		TrackDurationSeconds  uint32
		CompletionRatio       float32
	}

	err = conn.QueryRow(context.Background(),
		`SELECT user_uuid, music_uuid, artist_uuid, album_uuid, listen_duration_seconds, track_duration_seconds, completion_ratio 
		 FROM music_listen_events 
		 WHERE user_uuid = ? AND music_uuid = ?`,
		userUUID, musicUUID,
	).Scan(
		&result.UserUUID,
		&result.MusicUUID,
		&result.ArtistUUID,
		&result.AlbumUUID,
		&result.ListenDurationSeconds,
		&result.TrackDurationSeconds,
		&result.CompletionRatio,
	)

	require.NoError(t, err, "Failed to find inserted listen event")
	assert.Equal(t, userUUID, result.UserUUID)
	assert.Equal(t, musicUUID, result.MusicUUID)
	assert.Equal(t, artistUUID, result.ArtistUUID)
	assert.Equal(t, albumUUID, result.AlbumUUID)
	assert.Equal(t, uint32(120), result.ListenDurationSeconds)
	assert.Equal(t, uint32(180), result.TrackDurationSeconds)
	assert.InDelta(t, 0.67, result.CompletionRatio, 0.01)
}

func TestIntegration_LikeEvent_Ingestion(t *testing.T) {
	conn := SetupClickHouseForTest(t)
	handler := CreateTestHandler(t)

	userUUID := NewTestUUID()
	musicUUID := NewTestUUID()
	artistUUID := NewTestUUID()

	// Insert like event via handler
	reqBody := map[string]interface{}{
		"user_uuid":   userUUID,
		"music_uuid":  musicUUID,
		"artist_uuid": artistUUID,
	}

	req := createJSONRequest(t, "POST", "/events/like", reqBody)
	router := mux.NewRouter()
	router.HandleFunc("/events/like", handler.IngestLikeEvent).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["success"])

	// Verify data was inserted into ClickHouse
	var result struct {
		UserUUID   string
		MusicUUID  string
		ArtistUUID string
	}

	err = conn.QueryRow(context.Background(),
		`SELECT user_uuid, music_uuid, artist_uuid 
		 FROM music_like_events 
		 WHERE user_uuid = ? AND music_uuid = ?`,
		userUUID, musicUUID,
	).Scan(
		&result.UserUUID,
		&result.MusicUUID,
		&result.ArtistUUID,
	)

	require.NoError(t, err, "Failed to find inserted like event")
	assert.Equal(t, userUUID, result.UserUUID)
	assert.Equal(t, musicUUID, result.MusicUUID)
	assert.Equal(t, artistUUID, result.ArtistUUID)
}

func TestIntegration_ThemeEvent_Ingestion(t *testing.T) {
	conn := SetupClickHouseForTest(t)
	handler := CreateTestHandler(t)

	musicUUID := NewTestUUID()
	theme := "rock"

	// Insert theme event via handler
	reqBody := map[string]interface{}{
		"music_uuid": musicUUID,
		"theme":      theme,
	}

	req := createJSONRequest(t, "POST", "/events/theme", reqBody)
	router := mux.NewRouter()
	router.HandleFunc("/events/theme", handler.IngestThemeEvent).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["success"])

	// Verify data was inserted into ClickHouse
	var result struct {
		MusicUUID string
		Theme     string
	}

	err = conn.QueryRow(context.Background(),
		`SELECT music_uuid, theme 
		 FROM music_theme 
		 WHERE music_uuid = ?`,
		musicUUID,
	).Scan(
		&result.MusicUUID,
		&result.Theme,
	)

	require.NoError(t, err, "Failed to find inserted theme event")
	assert.Equal(t, musicUUID, result.MusicUUID)
	assert.Equal(t, theme, result.Theme)
}

func TestIntegration_UserDimEvent_Ingestion(t *testing.T) {
	conn := SetupClickHouseForTest(t)
	handler := CreateTestHandler(t)

	userUUID := NewTestUUID()
	country := "US"

	// Insert user dim event via handler
	reqBody := map[string]interface{}{
		"user_uuid":  userUUID,
		"country":    country,
		"created_at": "2024-01-01T00:00:00Z",
	}

	req := createJSONRequest(t, "POST", "/events/user", reqBody)
	router := mux.NewRouter()
	router.HandleFunc("/events/user", handler.IngestUserDimEvent).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["success"])

	// Verify data was inserted into ClickHouse
	var result struct {
		UserUUID string
		Country  string
	}

	err = conn.QueryRow(context.Background(),
		`SELECT user_uuid, country 
		 FROM user_dim 
		 WHERE user_uuid = ?`,
		userUUID,
	).Scan(
		&result.UserUUID,
		&result.Country,
	)

	require.NoError(t, err, "Failed to find inserted user dim event")
	assert.Equal(t, userUUID, result.UserUUID)
	assert.Equal(t, country, result.Country)
}

func TestIntegration_MultipleEvents_SameUser(t *testing.T) {
	conn := SetupClickHouseForTest(t)
	handler := CreateTestHandler(t)

	userUUID := NewTestUUID()
	musicUUID1 := NewTestUUID()
	musicUUID2 := NewTestUUID()
	artistUUID := NewTestUUID()
	albumUUID := NewTestUUID()

	// Insert first listen event
	reqBody1 := map[string]interface{}{
		"user_uuid":               userUUID,
		"music_uuid":              musicUUID1,
		"artist_uuid":             artistUUID,
		"album_uuid":              albumUUID,
		"listen_duration_seconds": 60,
		"track_duration_seconds":  180,
		"completion_ratio":        0.33,
	}

	req1 := createJSONRequest(t, "POST", "/events/listen", reqBody1)
	router := mux.NewRouter()
	router.HandleFunc("/events/listen", handler.IngestListenEvent).Methods("POST")

	rr1 := httptest.NewRecorder()
	router.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	ensureTimestampDistinct()

	// Insert second listen event
	reqBody2 := map[string]interface{}{
		"user_uuid":               userUUID,
		"music_uuid":              musicUUID2,
		"artist_uuid":             artistUUID,
		"album_uuid":              albumUUID,
		"listen_duration_seconds": 180,
		"track_duration_seconds":  180,
		"completion_ratio":        1.0,
	}

	req2 := createJSONRequest(t, "POST", "/events/listen", reqBody2)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	// Verify both events were inserted
	var count uint64
	err := conn.QueryRow(context.Background(),
		`SELECT count() FROM music_listen_events WHERE user_uuid = ?`,
		userUUID,
	).Scan(&count)

	require.NoError(t, err)
	assert.Equal(t, uint64(2), count, "Should have 2 listen events for user")
}

func TestIntegration_EventWithOptionalAlbumUUID(t *testing.T) {
	conn := SetupClickHouseForTest(t)
	handler := CreateTestHandler(t)

	userUUID := NewTestUUID()
	musicUUID := NewTestUUID()
	artistUUID := NewTestUUID()
	albumUUID := NewTestUUID()

	// Insert listen event with album_uuid
	reqBody := map[string]interface{}{
		"user_uuid":               userUUID,
		"music_uuid":              musicUUID,
		"artist_uuid":             artistUUID,
		"album_uuid":              albumUUID,
		"listen_duration_seconds": 120,
		"track_duration_seconds":  180,
		"completion_ratio":        0.67,
	}

	req := createJSONRequest(t, "POST", "/events/listen", reqBody)
	router := mux.NewRouter()
	router.HandleFunc("/events/listen", handler.IngestListenEvent).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify event was inserted
	var result struct {
		UserUUID  string
		MusicUUID string
		AlbumUUID *string
	}

	err := conn.QueryRow(context.Background(),
		`SELECT user_uuid, music_uuid, album_uuid 
		 FROM music_listen_events 
		 WHERE user_uuid = ? AND music_uuid = ?`,
		userUUID, musicUUID,
	).Scan(
		&result.UserUUID,
		&result.MusicUUID,
		&result.AlbumUUID,
	)

	require.NoError(t, err)
	require.NotNil(t, result.AlbumUUID, "Album UUID should not be nil")
	assert.Equal(t, albumUUID, *result.AlbumUUID)
}

// SetupClickHouseForTest creates a ClickHouse connection for integration tests
func SetupClickHouseForTest(t *testing.T) driver.Conn {
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
