//go:build e2e

package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTags_List tests listing all tags
func TestTags_List(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/tags", nil)
	require.NoError(t, err, "List tags request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestTags_GetByName tests getting a tag by name
func TestTags_GetByName(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/tags/pop", nil)
	require.NoError(t, err, "Get tag request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// Tag might or might not exist
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound,
		"Get tag should return OK or Not Found")
}

// TestTags_GetMusicByTag tests getting music by tag
func TestTags_GetMusicByTag(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/tags/pop/music", nil)
	require.NoError(t, err, "Get music by tag request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestHistory_GetUserHistory tests getting user listening history
func TestHistory_GetUserHistory(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/history", nil)
	require.NoError(t, err, "Get history request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestHistory_GetTopPlayed tests getting top played music
func TestHistory_GetTopPlayed(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/history/top", nil)
	require.NoError(t, err, "Get top played request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestSearch_Users tests searching for users
func TestSearch_Users(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/search/users?q=test", nil)
	require.NoError(t, err, "Search users request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestSearch_Artists tests searching for artists
func TestSearch_Artists(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/search/artists?q=test", nil)
	require.NoError(t, err, "Search artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestSearch_Albums tests searching for albums
func TestSearch_Albums(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/search/albums?q=test", nil)
	require.NoError(t, err, "Search albums request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestSearch_Music tests searching for music
func TestSearch_Music(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/search/music?q=test", nil)
	require.NoError(t, err, "Search music request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestSearch_Playlists tests searching for playlists
func TestSearch_Playlists(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/search/playlists?q=test", nil)
	require.NoError(t, err, "Search playlists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestSearch_EmptyQuery tests search with empty query
func TestSearch_EmptyQuery(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/search/music?q=", nil)
	require.NoError(t, err, "Search request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// Empty query might return bad request or empty results
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
		"Empty query should be handled")
}

// TestSearch_SpecialCharacters tests search with special characters
func TestSearch_SpecialCharacters(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/search/music?q=!@#$%^&*()", nil)
	require.NoError(t, err, "Search request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// Should handle gracefully
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
		"Special characters should be handled")
}

// TestPopular_PopularSongsAllTime tests getting all-time popular songs
func TestPopular_PopularSongsAllTime(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/popular/songs/all-time?limit=10", nil)
	require.NoError(t, err, "Get popular songs request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestPopular_PopularArtistsAllTime tests getting all-time popular artists
func TestPopular_PopularArtistsAllTime(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/popular/artists/all-time?limit=10", nil)
	require.NoError(t, err, "Get popular artists request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestPopular_PopularThemesAllTime tests getting all-time popular themes
func TestPopular_PopularThemesAllTime(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/popular/themes/all-time?limit=10", nil)
	require.NoError(t, err, "Get popular themes request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestPopular_PopularSongsByTheme tests getting popular songs by theme
func TestPopular_PopularSongsByTheme(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/popular/songs/theme/rock?limit=10", nil)
	require.NoError(t, err, "Get popular songs by theme request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestPopular_PopularSongsTimeframe tests getting popular songs by timeframe
func TestPopular_PopularSongsTimeframe(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/popular/songs/timeframe?start_date=2024-01-01&end_date=2024-12-31&limit=10", nil)
	require.NoError(t, err, "Get popular songs by timeframe request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// May return OK or Bad Request for invalid dates
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
		"Timeframe query should be handled")
}

// TestPopular_PopularArtistsTimeframe tests getting trending artists within date range
func TestPopular_PopularArtistsTimeframe(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/popular/artists/timeframe?start_date=2024-01-01&end_date=2024-12-31&limit=10", nil)
	require.NoError(t, err, "Get popular artists by timeframe request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// May return OK or Bad Request for invalid dates
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
		"Timeframe query should be handled")
}

// TestPopular_PopularThemesTimeframe tests getting trending themes within date range
func TestPopular_PopularThemesTimeframe(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/popular/themes/timeframe?start_date=2024-01-01&end_date=2024-12-31&limit=10", nil)
	require.NoError(t, err, "Get popular themes by timeframe request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// May return OK or Bad Request for invalid dates
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
		"Timeframe query should be handled")
}

// TestPopular_PopularSongsByThemeTimeframe tests getting trending songs for theme within date range
func TestPopular_PopularSongsByThemeTimeframe(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("GET", "/popular/songs/theme/rock/timeframe?start_date=2024-01-01&end_date=2024-12-31&limit=10", nil)
	require.NoError(t, err, "Get popular songs by theme and timeframe request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	// May return OK or Bad Request for invalid dates
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
		"Theme timeframe query should be handled")
}

// TestRecommend_Theme tests theme recommendation endpoint
func TestRecommend_Theme(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	resp, err := client.Request("POST", "/recommend/theme", map[string]interface{}{
		"user_uuid": client.GetUserUUID(),
	})

	require.NoError(t, err, "Recommend theme request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestEvents_Listen tests sending listen event
func TestEvents_Listen(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	eventData := map[string]interface{}{
		"user_uuid":               client.GetUserUUID(),
		"music_uuid":              NewTestUUID(),
		"artist_uuid":             NewTestUUID(),
		"album_uuid":              NewTestUUID(),
		"listen_duration_seconds": 120,
		"track_duration_seconds":  180,
		"completion_ratio":        0.67,
	}

	resp, err := client.Request("POST", "/events/listen", eventData)
	require.NoError(t, err, "Send listen event request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestEvents_Like tests sending like event
func TestEvents_Like(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	eventData := map[string]interface{}{
		"user_uuid":   client.GetUserUUID(),
		"music_uuid":  NewTestUUID(),
		"artist_uuid": NewTestUUID(),
	}

	resp, err := client.Request("POST", "/events/like", eventData)
	require.NoError(t, err, "Send like event request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestEvents_Theme tests sending theme event
func TestEvents_Theme(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	eventData := map[string]interface{}{
		"music_uuid": NewTestUUID(),
		"theme":      "rock",
	}

	resp, err := client.Request("POST", "/events/theme", eventData)
	require.NoError(t, err, "Send theme event request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestEvents_User tests sending user dimension event
func TestEvents_User(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	eventData := map[string]interface{}{
		"user_uuid": client.GetUserUUID(),
		"country":   "US",
	}

	resp, err := client.Request("POST", "/events/user", eventData)
	require.NoError(t, err, "Send user event request should not fail")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadGateway {
		t.Skip("Backend service not available")
	}

	AssertResponseStatus(t, resp, http.StatusOK)
}

// TestFiles_Public tests accessing public files
func TestFiles_Public(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Try to access a public file
	resp, err := client.RawRequest("GET", "/files/public/test.jpg", nil)
	require.NoError(t, err, "Get public file request should not fail")
	defer resp.Body.Close()

	// Should not return 404 (might return 403 or actual file)
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
}

// TestFiles_Private tests accessing private files without auth
func TestFiles_Private(t *testing.T) {
	config := GetTestConfig()
	client := NewTestClient(config.GatewayBaseURL)

	// Try to access a private file without auth
	resp, err := client.RawRequest("GET", "/files/private/test.jpg", nil)
	require.NoError(t, err, "Get private file request should not fail")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	AssertResponseStatus(t, resp, http.StatusUnauthorized)
}

// TestFiles_PrivateWithAuth tests accessing private files with auth
func TestFiles_PrivateWithAuth(t *testing.T) {
	config := GetTestConfig()
	client := SetupAuthenticatedClient(t, config)

	// Try to access a private file with auth
	resp, err := client.Request("GET", "/files/private/test.jpg", nil)
	require.NoError(t, err, "Get private file request should not fail")
	defer resp.Body.Close()

	// Should not return 401 (might return 404 or actual file)
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
}
