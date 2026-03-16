//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_PopularSongsAllTime(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data
	musicUUID1 := uuid.MustParse(NewTestUUID())
	musicUUID2 := uuid.MustParse(NewTestUUID())
	artistUUID := uuid.MustParse(NewTestUUID())

	// Insert two listen events with different durations
	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID1, artistUUID, 120)
	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID2, artistUUID, 60)

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	// Create request
	req := httptest.NewRequest("GET", "/popular/songs/all-time?limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/all-time", handler.PopularSongsAllTime).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	// Should have at least 2 songs
	assert.GreaterOrEqual(t, len(results), 2, "Should have at least 2 songs in results")
}

func TestIntegration_PopularArtistsAllTime(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data - same artist, different songs
	artistUUID := uuid.MustParse(NewTestUUID())
	musicUUID1 := uuid.MustParse(NewTestUUID())
	musicUUID2 := uuid.MustParse(NewTestUUID())

	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID1, artistUUID, 120)
	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID2, artistUUID, 90)

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	// Create request
	req := httptest.NewRequest("GET", "/popular/artists/all-time?limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/artists/all-time", handler.PopularArtistAllTime).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	// Should have at least 1 artist
	assert.GreaterOrEqual(t, len(results), 1, "Should have at least 1 artist in results")
}

func TestIntegration_PopularThemesAllTime(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data with themes
	musicUUID1 := uuid.MustParse(NewTestUUID())
	musicUUID2 := uuid.MustParse(NewTestUUID())
	artistUUID := uuid.MustParse(NewTestUUID())

	// Insert themes
	InsertTestTheme(t, conn, musicUUID1, "rock")
	InsertTestTheme(t, conn, musicUUID2, "pop")

	// Insert listen events
	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID1, artistUUID, 120)
	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID2, artistUUID, 90)

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	// Create request
	req := httptest.NewRequest("GET", "/popular/themes/all-time?limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/themes/all-time", handler.PopularThemeAllTime).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	// Should have at least 2 themes
	assert.GreaterOrEqual(t, len(results), 2, "Should have at least 2 themes in results")
}

func TestIntegration_PopularSongsByTheme(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data
	musicUUID := uuid.MustParse(NewTestUUID())
	artistUUID := uuid.MustParse(NewTestUUID())
	theme := "jazz"

	// Insert theme
	InsertTestTheme(t, conn, musicUUID, theme)

	// Insert listen event
	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID, artistUUID, 180)

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	// Create request
	req := httptest.NewRequest("GET", "/popular/songs/theme/jazz?limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/theme/{theme}", handler.PopularSongsAllTimeByTheme).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	// Should have at least 1 song for jazz theme
	assert.GreaterOrEqual(t, len(results), 1, "Should have at least 1 song for jazz theme")
}

func TestIntegration_PopularSongsTimeframe(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data
	musicUUID := uuid.MustParse(NewTestUUID())
	artistUUID := uuid.MustParse(NewTestUUID())

	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID, artistUUID, 120)

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Create request with timeframe
	req := httptest.NewRequest("GET", "/popular/songs/timeframe?start_date="+yesterday+"&end_date="+today+"&limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/timeframe", handler.PopularSongsTimeframe).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	// Should have at least 1 song
	assert.GreaterOrEqual(t, len(results), 1, "Should have at least 1 song in timeframe")
}

func TestIntegration_PopularArtistsTimeframe(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data
	artistUUID := uuid.MustParse(NewTestUUID())
	musicUUID := uuid.MustParse(NewTestUUID())

	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID, artistUUID, 90)

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Create request with timeframe
	req := httptest.NewRequest("GET", "/popular/artists/timeframe?start_date="+yesterday+"&end_date="+today+"&limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/artists/timeframe", handler.PopularArtistTimeframe).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	// Should have at least 1 artist
	assert.GreaterOrEqual(t, len(results), 1, "Should have at least 1 artist in timeframe")
}

func TestIntegration_PopularThemesTimeframe(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data
	musicUUID := uuid.MustParse(NewTestUUID())
	artistUUID := uuid.MustParse(NewTestUUID())

	InsertTestTheme(t, conn, musicUUID, "classical")
	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID, artistUUID, 200)

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Create request with timeframe
	req := httptest.NewRequest("GET", "/popular/themes/timeframe?start_date="+yesterday+"&end_date="+today+"&limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/themes/timeframe", handler.PopularThemeTimeframe).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	// Should have at least 1 theme
	assert.GreaterOrEqual(t, len(results), 1, "Should have at least 1 theme in timeframe")
}

func TestIntegration_PopularSongsTimeframeByTheme(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data
	musicUUID := uuid.MustParse(NewTestUUID())
	artistUUID := uuid.MustParse(NewTestUUID())
	theme := "electronic"

	InsertTestTheme(t, conn, musicUUID, theme)
	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID, artistUUID, 150)

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Create request with timeframe and theme
	req := httptest.NewRequest("GET", "/popular/songs/theme/electronic/timeframe?start_date="+yesterday+"&end_date="+today+"&limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/theme/{theme}/timeframe", handler.PopularSongsTimeframeByTheme).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	// Should have at least 1 song
	assert.GreaterOrEqual(t, len(results), 1, "Should have at least 1 song for electronic theme in timeframe")
}

func TestIntegration_EmptyDatabase(t *testing.T) {
	handler := CreateTestPopularityHandler(t)

	// Don't insert any data - test empty results

	// Create request
	req := httptest.NewRequest("GET", "/popular/songs/all-time?limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/all-time", handler.PopularSongsAllTime).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response - should return empty array
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	assert.Equal(t, 0, len(results), "Should return empty array when no data")
}

func TestIntegration_LimitParameter(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert multiple songs
	for i := 0; i < 5; i++ {
		musicUUID := uuid.MustParse(NewTestUUID())
		artistUUID := uuid.MustParse(NewTestUUID())
		InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID, artistUUID, uint32(60+i*10))
	}

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	// Test with limit=2
	req := httptest.NewRequest("GET", "/popular/songs/all-time?limit=2", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/all-time", handler.PopularSongsAllTime).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	// Should respect the limit
	assert.LessOrEqual(t, len(results), 2, "Should respect the limit parameter")
}

func TestIntegration_InvalidTimeframe(t *testing.T) {
	handler := CreateTestPopularityHandler(t)

	// Test with invalid date range (end before start)
	req := httptest.NewRequest("GET", "/popular/songs/timeframe?start_date=2024-12-31&end_date=2024-01-01&limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/timeframe", handler.PopularSongsTimeframe).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response - should return error
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestIntegration_InvalidDateFormat(t *testing.T) {
	handler := CreateTestPopularityHandler(t)

	// Test with invalid date format
	req := httptest.NewRequest("GET", "/popular/songs/timeframe?start_date=invalid&end_date=invalid&limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/timeframe", handler.PopularSongsTimeframe).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response - should return error
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestIntegration_MissingDateParameters(t *testing.T) {
	handler := CreateTestPopularityHandler(t)

	// Test without required date parameters
	req := httptest.NewRequest("GET", "/popular/songs/timeframe?limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/timeframe", handler.PopularSongsTimeframe).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response - should return error
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestIntegration_DefaultLimit(t *testing.T) {
	handler := CreateTestPopularityHandler(t)

	// Test without limit parameter - should use default of 50
	req := httptest.NewRequest("GET", "/popular/songs/all-time", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/all-time", handler.PopularSongsAllTime).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response - should work fine
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	// Should return empty or results up to default limit
	assert.True(t, len(results) <= 50, "Should not exceed default limit of 50")
}

func TestIntegration_InvalidLimitValues(t *testing.T) {
	handler := CreateTestPopularityHandler(t)

	testCases := []struct {
		name     string
		limit    string
		expected int // max expected results
	}{
		{"negative_limit", "-1", 50},
		{"zero_limit", "0", 50},
		{"over_100_limit", "150", 50},
		{"non_numeric_limit", "abc", 50},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/popular/songs/all-time?limit="+tc.limit, nil)
			rr := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/popular/songs/all-time", handler.PopularSongsAllTime).Methods("GET")
			router.ServeHTTP(rr, req)

			// Should still return OK with default limit
			assert.Equal(t, http.StatusOK, rr.Code)

			var results []map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &results)
			require.NoError(t, err)
			assert.True(t, len(results) <= tc.expected, "Should use default limit for invalid: %s", tc.limit)
		})
	}
}

func TestIntegration_ResponseStructure(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data
	musicUUID := uuid.MustParse(NewTestUUID())
	artistUUID := uuid.MustParse(NewTestUUID())

	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID, artistUUID, 120)

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	// Test song response structure
	req := httptest.NewRequest("GET", "/popular/songs/all-time?limit=1", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/all-time", handler.PopularSongsAllTime).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	if len(results) > 0 {
		song := results[0]
		// Verify expected fields exist
		assert.Contains(t, song, "music_uuid", "Response should contain music_uuid")
		assert.Contains(t, song, "decay_plays", "Response should contain decay_plays")
		assert.Contains(t, song, "decay_listen_seconds", "Response should contain decay_listen_seconds")
	}

	// Test artist response structure
	req = httptest.NewRequest("GET", "/popular/artists/all-time?limit=1", nil)
	rr = httptest.NewRecorder()

	router = mux.NewRouter()
	router.HandleFunc("/popular/artists/all-time", handler.PopularArtistAllTime).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	err = json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	if len(results) > 0 {
		artist := results[0]
		assert.Contains(t, artist, "artist_uuid", "Response should contain artist_uuid")
		assert.Contains(t, artist, "decay_plays", "Response should contain decay_plays")
		assert.Contains(t, artist, "decay_listen_seconds", "Response should contain decay_listen_seconds")
	}

	// Test theme response structure
	req = httptest.NewRequest("GET", "/popular/themes/all-time?limit=1", nil)
	rr = httptest.NewRecorder()

	router = mux.NewRouter()
	router.HandleFunc("/popular/themes/all-time", handler.PopularThemeAllTime).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	err = json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	if len(results) > 0 {
		theme := results[0]
		assert.Contains(t, theme, "theme", "Response should contain theme")
		assert.Contains(t, theme, "decay_plays", "Response should contain decay_plays")
		assert.Contains(t, theme, "decay_listen_seconds", "Response should contain decay_listen_seconds")
	}
}

func TestIntegration_PopularSongsTimeframeResponseStructure(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data
	musicUUID := uuid.MustParse(NewTestUUID())
	artistUUID := uuid.MustParse(NewTestUUID())

	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID, artistUUID, 120)

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Test timeframe response structure
	req := httptest.NewRequest("GET", "/popular/songs/timeframe?start_date="+yesterday+"&end_date="+today+"&limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/timeframe", handler.PopularSongsTimeframe).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	if len(results) > 0 {
		song := results[0]
		assert.Contains(t, song, "music_uuid", "Response should contain music_uuid")
		assert.Contains(t, song, "plays", "Response should contain plays")
		assert.Contains(t, song, "listen_seconds", "Response should contain listen_seconds")
	}
}

func TestIntegration_ThemeNotFound(t *testing.T) {
	handler := CreateTestPopularityHandler(t)

	// Test with non-existent theme
	req := httptest.NewRequest("GET", "/popular/songs/theme/nonexistent_theme?limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/theme/{theme}", handler.PopularSongsAllTimeByTheme).Methods("GET")
	router.ServeHTTP(rr, req)

	// Should return OK with empty results
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	assert.Equal(t, 0, len(results), "Should return empty for non-existent theme")
}

func TestIntegration_PopularSongsTimeframeByThemeResponseStructure(t *testing.T) {
	conn := SetupClickHouse(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data
	musicUUID := uuid.MustParse(NewTestUUID())
	artistUUID := uuid.MustParse(NewTestUUID())
	theme := "blues"

	InsertTestTheme(t, conn, musicUUID, theme)
	InsertTestListenEvent(t, conn, uuid.MustParse(NewTestUUID()), musicUUID, artistUUID, 150)

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Test timeframe by theme response structure
	req := httptest.NewRequest("GET", "/popular/songs/theme/"+theme+"/timeframe?start_date="+yesterday+"&end_date="+today+"&limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/theme/{theme}/timeframe", handler.PopularSongsTimeframeByTheme).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	if len(results) > 0 {
		song := results[0]
		assert.Contains(t, song, "music_uuid", "Response should contain music_uuid")
		assert.Contains(t, song, "theme", "Response should contain theme")
		assert.Contains(t, song, "plays", "Response should contain plays")
		assert.Contains(t, song, "listen_seconds", "Response should contain listen_seconds")
	}
}

// TestIntegration_PopularSongsAllTimeDirectQuery tests the handler by querying directly
func TestIntegration_PopularSongsAllTimeDirectQuery(t *testing.T) {
	db := GetClickHouseDB(t)
	handler := CreateTestPopularityHandler(t)

	// Insert test data
	musicUUID := uuid.MustParse(NewTestUUID())
	artistUUID := uuid.MustParse(NewTestUUID())

	// Insert directly via SQL
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO music_listen_events (user_uuid, music_uuid, artist_uuid, listen_duration_seconds, track_duration_seconds, completion_ratio)
		VALUES (?, ?, ?, ?, ?, ?)
	`, uuid.MustParse(NewTestUUID()), musicUUID, artistUUID, 100, 180, 0.55)

	require.NoError(t, err, "Failed to insert test data")

	// Wait for materialized views to update
	time.Sleep(500 * time.Millisecond)

	// Use the handler
	req := httptest.NewRequest("GET", "/popular/songs/all-time?limit=10", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/popular/songs/all-time", handler.PopularSongsAllTime).Methods("GET")
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var results []map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &results)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(results), 1, "Should have at least 1 result")
}
