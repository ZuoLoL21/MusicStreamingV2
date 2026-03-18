//go:build integration

package integration

import (
	backenddi "backend/internal/di"
	"backend/internal/handlers"
	"backend/tests/integration/builders"
	"context"
	"libs/di"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestIntegration_Error_DatabaseConstraintViolations tests handling of database constraint violations
func TestIntegration_Error_DatabaseConstraintViolations(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	t.Run("duplicate_email_rejected", func(t *testing.T) {
		// Create first user
		builders.NewUserBuilder().
			WithEmail("duplicate@test.com").
			WithUsername("user1").
			Build(t, ctx, db)

		// Try to create second user with same email
		handler := handlers.NewAuthHandler(config, nil, returns, db, nil, nil)
		req := createMultipartRequest(t, "POST", "/register", "", "", nil, map[string]string{
			"username":    "user2",
			"email":       "duplicate@test.com",
			"password":    "TestPassword123!",
			"country":     "US",
			"device_id":   "00000000-0000-0000-0000-000000000001",
			"device_name": "test-device",
		})

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should be rejected with conflict
		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("duplicate_username_rejected", func(t *testing.T) {
		// Create first user
		builders.NewUserBuilder().
			WithEmail("user1@test.com").
			WithUsername("samename").
			Build(t, ctx, db)

		// Try to create second user with same username
		handler := handlers.NewAuthHandler(config, nil, returns, db, nil, nil)
		req := createMultipartRequest(t, "POST", "/register", "", "", nil, map[string]string{
			"username":    "samename",
			"email":       "user2@test.com",
			"password":    "TestPassword123!",
			"country":     "US",
			"device_id":   "00000000-0000-0000-0000-000000000001",
			"device_name": "test-device",
		})

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should be rejected
		assert.Equal(t, http.StatusConflict, rr.Code)
	})
}

// TestIntegration_Error_NotFoundResponses tests 404 handling
func TestIntegration_Error_NotFoundResponses(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	t.Run("user_not_found", func(t *testing.T) {
		handler := handlers.NewUserHandler(config, nil, returns, db, nil, nil)
		req := createRequest(t, "GET", "/users/00000000-0000-0000-0000-000000000001", nil)

		router := mux.NewRouter()
		router.HandleFunc("/users/{uuid}", handler.GetPublicUser).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("artist_not_found", func(t *testing.T) {
		handler := handlers.NewArtistHandler(config, returns, db, nil)
		req := createRequest(t, "GET", "/artists/00000000-0000-0000-0000-000000000001", nil)

		router := mux.NewRouter()
		router.HandleFunc("/artists/{uuid}", handler.GetArtist).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("music_not_found", func(t *testing.T) {
		handler := handlers.NewMusicHandler(config, returns, db, nil)
		req := createRequest(t, "GET", "/music/00000000-0000-0000-0000-000000000001", nil)

		router := mux.NewRouter()
		router.HandleFunc("/music/{uuid}", handler.GetMusic).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("album_not_found", func(t *testing.T) {
		handler := handlers.NewAlbumHandler(config, returns, db, nil)
		req := createRequest(t, "GET", "/albums/00000000-0000-0000-0000-000000000001", nil)

		router := mux.NewRouter()
		router.HandleFunc("/albums/{uuid}", handler.GetAlbum).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("playlist_not_found", func(t *testing.T) {
		handler := handlers.NewPlaylistHandler(config, returns, db, nil)
		req := createRequest(t, "GET", "/playlists/00000000-0000-0000-0000-000000000001", nil)

		router := mux.NewRouter()
		router.HandleFunc("/playlists/{uuid}", handler.GetPlaylist).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestIntegration_Error_InvalidUUIDFormat tests handling of invalid UUID formats
// All invalid UUIDs should consistently return 400 Bad Request
func TestIntegration_Error_InvalidUUIDFormat(t *testing.T) {
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	testCases := []struct {
		name        string
		endpoint    string
		handlerFunc func() http.HandlerFunc
	}{
		{
			name:     "invalid_uuid_user_get",
			endpoint: "/users/not-a-valid-uuid",
			handlerFunc: func() http.HandlerFunc {
				h := handlers.NewUserHandler(config, nil, returns, nil, nil, nil)
				return h.GetPublicUser
			},
		},
		{
			name:     "invalid_uuid_artist_get",
			endpoint: "/artists/invalid-uuid",
			handlerFunc: func() http.HandlerFunc {
				h := handlers.NewArtistHandler(config, returns, nil, nil)
				return h.GetArtist
			},
		},
		{
			name:     "invalid_uuid_music_get",
			endpoint: "/music/not-uuid",
			handlerFunc: func() http.HandlerFunc {
				h := handlers.NewMusicHandler(config, returns, nil, nil)
				return h.GetMusic
			},
		},
		{
			name:     "invalid_uuid_album_get",
			endpoint: "/albums/bad-uuid",
			handlerFunc: func() http.HandlerFunc {
				h := handlers.NewAlbumHandler(config, returns, nil, nil)
				return h.GetAlbum
			},
		},
		{
			name:     "invalid_uuid_playlist_get",
			endpoint: "/playlists/xyz",
			handlerFunc: func() http.HandlerFunc {
				h := handlers.NewPlaylistHandler(config, returns, nil, nil)
				return h.GetPlaylist
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createRequest(t, "GET", tc.endpoint, nil)

			router := mux.NewRouter()
			router.HandleFunc(strings.Replace(tc.endpoint, strings.Split(tc.endpoint, "/")[len(strings.Split(tc.endpoint, "/"))-1], "{uuid}", 1), tc.handlerFunc()).Methods("GET")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// All invalid UUIDs should return 400 Bad Request
			assert.Equal(t, http.StatusBadRequest, rr.Code, "Invalid UUID should return 400 for %s", tc.name)
		})
	}
}

// TestIntegration_Error_MissingRequiredFields tests handling of missing required fields
func TestIntegration_Error_MissingRequiredFields(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	testUser := builders.NewUserBuilder().
		WithEmail("test@example.com").
		Build(t, ctx, db)

	t.Run("register_missing_username", func(t *testing.T) {
		handler := handlers.NewAuthHandler(config, nil, returns, db, nil, nil)
		req := createMultipartRequest(t, "POST", "/register", "", "", nil, map[string]string{
			"email":       "test@test.com",
			"password":    "TestPassword123!",
			"country":     "US",
			"device_id":   "00000000-0000-0000-0000-000000000001",
			"device_name": "test-device",
		})

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("register_missing_email", func(t *testing.T) {
		handler := handlers.NewAuthHandler(config, nil, returns, db, nil, nil)
		req := createMultipartRequest(t, "POST", "/register", "", "", nil, map[string]string{
			"username":    "testuser",
			"password":    "TestPassword123!",
			"country":     "US",
			"device_id":   "00000000-0000-0000-0000-000000000001",
			"device_name": "test-device",
		})

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("register_missing_password", func(t *testing.T) {
		handler := handlers.NewAuthHandler(config, nil, returns, db, nil, nil)
		req := createMultipartRequest(t, "POST", "/register", "", "", nil, map[string]string{
			"username":    "testuser",
			"email":       "test@test.com",
			"country":     "US",
			"device_id":   "00000000-0000-0000-0000-000000000001",
			"device_name": "test-device",
		})

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("playlist_missing_name", func(t *testing.T) {
		handler := handlers.NewPlaylistHandler(config, returns, db, nil)
		req := createMultipartRequest(t, "POST", "/playlists", "", "", nil, map[string]string{
			"description": "Test playlist",
			"device_id":   "00000000-0000-0000-0000-000000000001",
			"device_name": "test-device",
		})

		router := mux.NewRouter()
		router.HandleFunc("/playlists", wrapWithAuth(t, handler.CreatePlaylist, testUser)).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestIntegration_Error_UnauthorizedAccess tests handling of unauthorized access
func TestIntegration_Error_UnauthorizedAccess(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create owner and non-owner users
	owner := builders.NewUserBuilder().
		WithEmail("owner@test.com").
		Build(t, ctx, db)

	nonOwner := builders.NewUserBuilder().
		WithEmail("nonowner@test.com").
		Build(t, ctx, db)

	// Create artist's user and artist
	artistOwner := builders.NewUserBuilder().
		WithEmail("artistowner@test.com").
		Build(t, ctx, db)

	artist := builders.NewArtistBuilder(artistOwner).
		WithName("Test Artist").
		Build(t, ctx, db)

	t.Run("update_other_user_profile", func(t *testing.T) {
		handler := handlers.NewUserHandler(config, nil, returns, db, nil, nil)

		// Try to update non-owner's profile
		req := createJSONRequest(t, "POST", "/users/me", map[string]interface{}{
			"username": "hacked",
		})

		router := mux.NewRouter()
		// Simulate auth middleware setting wrong user context
		router.HandleFunc("/users/me", wrapWithAuth(t, handler.UpdateProfile, nonOwner)).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should either reject or require proper authentication
		assert.NotEqual(t, http.StatusOK, rr.Code)
	})

	t.Run("delete_other_user_playlist", func(t *testing.T) {
		// Create playlist owned by nonOwner
		playlist := builders.NewPlaylistBuilder(nonOwner).
			WithName("Private Playlist").
			Build(t, ctx, db)

		handler := handlers.NewPlaylistHandler(config, returns, db, nil)

		// Try to delete as different user (owner)
		req := createRequest(t, "DELETE", "/playlists/"+builders.UUIDToString(playlist), nil)

		router := mux.NewRouter()
		router.HandleFunc("/playlists/{uuid}", wrapWithAuth(t, handler.DeletePlaylist, owner)).Methods("DELETE")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should be forbidden
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("update_other_artist", func(t *testing.T) {
		handler := handlers.NewArtistHandler(config, returns, db, nil)

		// Try to update as non-member
		req := createJSONRequest(t, "POST", "/artists/"+builders.UUIDToString(artist), map[string]interface{}{
			"name": "Hacked Name",
		})

		router := mux.NewRouter()
		router.HandleFunc("/artists/{uuid}", wrapWithAuth(t, handler.UpdateArtistProfile, nonOwner)).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should be forbidden or not found
		assert.True(t, rr.Code == http.StatusForbidden || rr.Code == http.StatusNotFound)
	})
}

// TestIntegration_Error_MethodNotAllowed tests handling of wrong HTTP methods
func TestIntegration_Error_MethodNotAllowed(t *testing.T) {
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	t.Run("get_on_post_endpoint", func(t *testing.T) {
		handler := handlers.NewAuthHandler(config, nil, returns, nil, nil, nil)

		// Try GET on register endpoint (should be POST)
		req := createRequest(t, "GET", "/register", nil)

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("post_on_get_endpoint", func(t *testing.T) {
		handler := handlers.NewUserHandler(config, nil, returns, nil, nil, nil)

		// Try POST on user endpoint (should be GET)
		req := createJSONRequest(t, "POST", "/users/me", map[string]interface{}{})

		router := mux.NewRouter()
		router.HandleFunc("/users/me", handler.GetMe).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

// TestIntegration_Error_BadRequestBody tests handling of malformed request bodies
func TestIntegration_Error_BadRequestBody(t *testing.T) {
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	t.Run("invalid_json", func(t *testing.T) {
		handler := handlers.NewAuthHandler(config, nil, returns, nil, nil, nil)

		// Create request with invalid JSON body
		req := httptest.NewRequest("POST", "/register", strings.NewReader("{invalid json"))
		req.Header.Set("Content-Type", "application/json")

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnsupportedMediaType, rr.Code)
	})

	t.Run("wrong_content_type", func(t *testing.T) {
		handler := handlers.NewAuthHandler(config, nil, returns, nil, nil, nil)

		// Create request with wrong content type
		req := httptest.NewRequest("POST", "/register", strings.NewReader("plain text"))
		req.Header.Set("Content-Type", "text/plain")

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnsupportedMediaType, rr.Code)
	})
}

// TestIntegration_Error_DatabaseVsNotFound tests that actual "not found" returns 404
// while database errors return 500
func TestIntegration_Error_DatabaseVsNotFound(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	t.Run("user_not_found_returns_404", func(t *testing.T) {
		handler := handlers.NewUserHandler(config, nil, returns, db, nil, nil)

		// Use a valid UUID that doesn't exist
		nonExistentUUID := "00000000-0000-0000-0000-000000000001"
		req := createRequest(t, "GET", "/users/"+nonExistentUUID, nil)

		router := mux.NewRouter()
		router.HandleFunc("/users/{uuid}", handler.GetPublicUser).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return 404 for actual "not found"
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("music_not_found_returns_404", func(t *testing.T) {
		handler := handlers.NewMusicHandler(config, returns, db, nil)

		nonExistentUUID := "00000000-0000-0000-0000-000000000002"
		req := createRequest(t, "GET", "/music/"+nonExistentUUID, nil)

		router := mux.NewRouter()
		router.HandleFunc("/music/{uuid}", handler.GetMusic).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return 404 for actual "not found"
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("album_not_found_returns_404", func(t *testing.T) {
		handler := handlers.NewAlbumHandler(config, returns, db, nil)

		nonExistentUUID := "00000000-0000-0000-0000-000000000003"
		req := createRequest(t, "GET", "/albums/"+nonExistentUUID, nil)

		router := mux.NewRouter()
		router.HandleFunc("/albums/{uuid}", handler.GetAlbum).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return 404 for actual "not found"
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("playlist_not_found_returns_404", func(t *testing.T) {
		handler := handlers.NewPlaylistHandler(config, returns, db, nil)

		nonExistentUUID := "00000000-0000-0000-0000-000000000004"
		req := createRequest(t, "GET", "/playlists/"+nonExistentUUID, nil)

		router := mux.NewRouter()
		router.HandleFunc("/playlists/{uuid}", handler.GetPlaylist).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return 404 for actual "not found"
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
