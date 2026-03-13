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
		handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)
		req := createJSONRequest(t, "POST", "/register", map[string]interface{}{
			"username": "user2",
			"email":    "duplicate@test.com",
			"password": "TestPassword123!",
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
		handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)
		req := createJSONRequest(t, "POST", "/register", map[string]interface{}{
			"username": "samename",
			"email":    "user2@test.com",
			"password": "TestPassword123!",
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
		handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)
		req := createRequest(t, "GET", "/users/00000000-0000-0000-0000-000000000001", nil)

		router := mux.NewRouter()
		router.HandleFunc("/users/{uuid}", handler.GetPublicUser).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("artist_not_found", func(t *testing.T) {
		handler := handlers.NewArtistHandler(logger, config, returns, db, nil)
		req := createRequest(t, "GET", "/artists/00000000-0000-0000-0000-000000000001", nil)

		router := mux.NewRouter()
		router.HandleFunc("/artists/{uuid}", handler.GetArtist).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("music_not_found", func(t *testing.T) {
		handler := handlers.NewMusicHandler(logger, config, returns, db, nil)
		req := createRequest(t, "GET", "/music/00000000-0000-0000-0000-000000000001", nil)

		router := mux.NewRouter()
		router.HandleFunc("/music/{uuid}", handler.GetMusic).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("album_not_found", func(t *testing.T) {
		handler := handlers.NewAlbumHandler(logger, config, returns, db, nil)
		req := createRequest(t, "GET", "/albums/00000000-0000-0000-0000-000000000001", nil)

		router := mux.NewRouter()
		router.HandleFunc("/albums/{uuid}", handler.GetAlbum).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("playlist_not_found", func(t *testing.T) {
		handler := handlers.NewPlaylistHandler(logger, config, returns, db, nil)
		req := createRequest(t, "GET", "/playlists/00000000-0000-0000-0000-000000000001", nil)

		router := mux.NewRouter()
		router.HandleFunc("/playlists/{uuid}", handler.GetPlaylist).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestIntegration_Error_InvalidUUIDFormat tests handling of invalid UUID formats
func TestIntegration_Error_InvalidUUIDFormat(t *testing.T) {
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	t.Run("invalid_uuid_user_get", func(t *testing.T) {
		handler := handlers.NewUserHandler(logger, config, nil, returns, nil, nil)
		req := createRequest(t, "GET", "/users/not-a-valid-uuid", nil)

		router := mux.NewRouter()
		router.HandleFunc("/users/{uuid}", handler.GetPublicUser).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid_uuid_artist_get", func(t *testing.T) {
		handler := handlers.NewArtistHandler(logger, config, returns, nil, nil)
		req := createRequest(t, "GET", "/artists/invalid-uuid", nil)

		router := mux.NewRouter()
		router.HandleFunc("/artists/{uuid}", handler.GetArtist).Methods("GET")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
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
		handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)
		req := createJSONRequest(t, "POST", "/register", map[string]interface{}{
			"email":    "test@test.com",
			"password": "TestPassword123!",
		})

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("register_missing_email", func(t *testing.T) {
		handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)
		req := createJSONRequest(t, "POST", "/register", map[string]interface{}{
			"username": "testuser",
			"password": "TestPassword123!",
		})

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("register_missing_password", func(t *testing.T) {
		handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)
		req := createJSONRequest(t, "POST", "/register", map[string]interface{}{
			"username": "testuser",
			"email":    "test@test.com",
		})

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("playlist_missing_name", func(t *testing.T) {
		handler := handlers.NewPlaylistHandler(logger, config, returns, db, nil)
		req := createJSONRequest(t, "POST", "/playlists", map[string]interface{}{
			"description": "Test playlist",
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
		handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)

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

		handler := handlers.NewPlaylistHandler(logger, config, returns, db, nil)

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
		handler := handlers.NewArtistHandler(logger, config, returns, db, nil)

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
		handler := handlers.NewUserHandler(logger, config, nil, returns, nil, nil)

		// Try GET on register endpoint (should be POST)
		req := createRequest(t, "GET", "/register", nil)

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("post_on_get_endpoint", func(t *testing.T) {
		handler := handlers.NewUserHandler(logger, config, nil, returns, nil, nil)

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
		handler := handlers.NewUserHandler(logger, config, nil, returns, nil, nil)

		// Create request with invalid JSON body
		req := httptest.NewRequest("POST", "/register", strings.NewReader("{invalid json"))
		req.Header.Set("Content-Type", "application/json")

		router := mux.NewRouter()
		router.HandleFunc("/register", handler.Register).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("wrong_content_type", func(t *testing.T) {
		handler := handlers.NewUserHandler(logger, config, nil, returns, nil, nil)

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

// Helper function for creating request without body
func createRequestWithoutBody(t *testing.T, method, path string) *http.Request {
	t.Helper()
	return httptest.NewRequest(method, path, nil)
}
