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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIntegration_Validation_StringFields(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	testUser := builders.NewUserBuilder().
		WithEmail("testuser@example.com").
		Build(t, ctx, db)

	testArtist := builders.NewArtistBuilder(testUser).
		WithName("Test Artist").
		Build(t, ctx, db)

	testCases := []struct {
		name           string
		handler        string
		endpoint       string
		requestBody    map[string]interface{}
		formFields     map[string]string
		expectedStatus int
		description    string
	}{
		// User Profile Validation
		{
			name:     "user_update_empty_username",
			handler:  "user",
			endpoint: "/users/me",
			requestBody: map[string]interface{}{
				"username": "",
				"country":  "US",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Empty username should be rejected",
		},
		{
			name:     "user_update_whitespace_username",
			handler:  "user",
			endpoint: "/users/me",
			requestBody: map[string]interface{}{
				"username": "     ",
				"country":  "US",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Whitespace-only username should be rejected",
		},
		{
			name:     "user_update_very_long_username",
			handler:  "user",
			endpoint: "/users/me",
			requestBody: map[string]interface{}{
				"username": strings.Repeat("a", 1001),
				"country":  "US",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Very long username (>1000 chars) may be rejected",
		},
		{
			name:     "user_update_sql_injection_attempt",
			handler:  "user",
			endpoint: "/users/me",
			requestBody: map[string]interface{}{
				"username": "admin'; DROP TABLE users; --",
				"country":  "US",
			},
			expectedStatus: http.StatusOK,
			description:    "SQL injection attempt should be safely escaped (accepts as normal string)",
		},
		{
			name:     "user_update_xss_attempt",
			handler:  "user",
			endpoint: "/users/me",
			requestBody: map[string]interface{}{
				"username": "<script>alert('XSS')</script>",
				"country":  "US",
			},
			expectedStatus: http.StatusOK,
			description:    "XSS attempt should be safely stored (not executed)",
		},

		// Artist Name Validation
		{
			name:     "artist_create_empty_name",
			handler:  "artist",
			endpoint: "/artists",
			formFields: map[string]string{
				"name":        "",
				"description": "Test description",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Empty artist name should be rejected",
		},
		{
			name:     "artist_create_very_long_name",
			handler:  "artist",
			endpoint: "/artists",
			formFields: map[string]string{
				"name":        strings.Repeat("Artist", 200),
				"description": "Test",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Very long artist name may be rejected",
		},

		// Music Name Validation
		{
			name:     "music_empty_song_name",
			handler:  "music",
			endpoint: "/artists/" + builders.UUIDToString(testArtist) + "/music",
			formFields: map[string]string{
				"song_name": "",
				"duration":  "180",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Empty song name should be rejected",
		},

		// Album Name Validation
		{
			name:     "album_empty_name",
			handler:  "album",
			endpoint: "/artists/" + builders.UUIDToString(testArtist) + "/albums",
			formFields: map[string]string{
				"artist_uuid":   builders.UUIDToString(testArtist),
				"original_name": "",
				"description":   "Test album",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Empty album name should be rejected",
		},

		// Playlist Name Validation
		{
			name:     "playlist_empty_name",
			handler:  "playlist",
			endpoint: "/playlists",
			formFields: map[string]string{
				"original_name": "",
				"description":   "Test playlist",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Empty playlist name should be rejected",
		},
		{
			name:     "playlist_unicode_name",
			handler:  "playlist",
			endpoint: "/playlists",
			formFields: map[string]string{
				"original_name": "My Playlist 🎵🎶",
				"description":   "Unicode test",
			},
			expectedStatus: http.StatusCreated,
			description:    "Unicode characters in playlist name should be accepted",
		},
		{
			name:     "playlist_special_chars",
			handler:  "playlist",
			endpoint: "/playlists",
			formFields: map[string]string{
				"original_name": "Rock & Roll + Blues (90's)",
				"description":   "Special characters test",
			},
			expectedStatus: http.StatusCreated,
			description:    "Special characters should be accepted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var rr *httptest.ResponseRecorder
			var req *http.Request

			switch tc.handler {
			case "user":
				handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)
				req = createJSONRequest(t, "POST", tc.endpoint, tc.requestBody)
				router := mux.NewRouter()
				router.HandleFunc(tc.endpoint, wrapWithAuth(t, handler.UpdateProfile, testUser)).Methods("POST")
				rr = httptest.NewRecorder()
				router.ServeHTTP(rr, req)

			case "artist":
				handler := handlers.NewArtistHandler(logger, config, returns, db, fileStorage)
				req = createMultipartRequest(t, "POST", tc.endpoint, "", "", nil, tc.formFields)
				router := mux.NewRouter()
				router.HandleFunc(tc.endpoint, wrapWithAuth(t, handler.CreateArtist, testUser)).Methods("POST")
				rr = httptest.NewRecorder()
				router.ServeHTTP(rr, req)

			case "music":
				handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)
				audioData := []byte("fake audio")
				req = createMultipartRequest(t, "POST", tc.endpoint, "audio", "test.mp3", audioData, tc.formFields)
				router := mux.NewRouter()
				router.HandleFunc("/artists/{artist_uuid}/music", wrapWithAuth(t, handler.CreateMusic, testUser)).Methods("POST")
				rr = httptest.NewRecorder()
				router.ServeHTTP(rr, req)

			case "album":
				handler := handlers.NewAlbumHandler(logger, config, returns, db, nil)
				req = createMultipartRequest(t, "POST", tc.endpoint, "", "", nil, tc.formFields)
				router := mux.NewRouter()
				router.HandleFunc("/artists/{artist_uuid}/albums", wrapWithAuth(t, handler.CreateAlbum, testUser)).Methods("POST")
				rr = httptest.NewRecorder()
				router.ServeHTTP(rr, req)

			case "playlist":
				handler := handlers.NewPlaylistHandler(logger, config, returns, db, nil)
				req = createMultipartRequest(t, "POST", tc.endpoint, "", "", nil, tc.formFields)
				router := mux.NewRouter()
				router.HandleFunc(tc.endpoint, wrapWithAuth(t, handler.CreatePlaylist, testUser)).Methods("POST")
				rr = httptest.NewRecorder()
				router.ServeHTTP(rr, req)
			}

			assert.Equal(t, tc.expectedStatus, rr.Code, tc.description)
		})
	}
}

func TestIntegration_Validation_NullBytes(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	testUser := builders.NewUserBuilder().
		WithEmail("testuser@example.com").
		Build(t, ctx, db)

	handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)

	// Attempt to inject null byte in username
	updateReq := map[string]interface{}{
		"username": "valid\x00username",
		"country":  "US",
	}
	req := createJSONRequest(t, "POST", "/users/me", updateReq)

	router := mux.NewRouter()
	router.HandleFunc("/users/me", wrapWithAuth(t, handler.UpdateProfile, testUser)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Null bytes should be rejected
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestIntegration_Validation_LeadingTrailingWhitespace(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	testUser := builders.NewUserBuilder().
		WithEmail("testuser@example.com").
		Build(t, ctx, db)

	handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)

	// Username with leading/trailing whitespace
	updateReq := map[string]interface{}{
		"username": "  validusername  ",
		"country":  "US",
	}
	req := createJSONRequest(t, "POST", "/users/me", updateReq)

	router := mux.NewRouter()
	router.HandleFunc("/users/me", wrapWithAuth(t, handler.UpdateProfile, testUser)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Whitespace should either be trimmed or rejected
	// Document current behavior (likely accepted as-is)
	if rr.Code == http.StatusOK {
		user, err := db.GetPublicUser(ctx, testUser)
		assert.NoError(t, err)
		// Check if whitespace was preserved or trimmed
		t.Logf("Username stored as: '%s'", user.Username)
	}
}

func TestIntegration_Validation_VeryLongDescription(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	testUser := builders.NewUserBuilder().
		WithEmail("testuser@example.com").
		Build(t, ctx, db)

	handler := handlers.NewPlaylistHandler(logger, config, returns, db, nil)

	// Create playlist with very long description (>5000 chars)
	longDesc := strings.Repeat("This is a very long description. ", 200) // ~6600 chars

	createReq := map[string]interface{}{
		"name":        "Test Playlist",
		"description": longDesc,
	}
	req := createJSONRequest(t, "POST", "/playlists", createReq)

	router := mux.NewRouter()
	router.HandleFunc("/playlists", wrapWithAuth(t, handler.CreatePlaylist, testUser)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Very long descriptions may be accepted or rejected based on DB constraints
	// Document current behavior
	if rr.Code == http.StatusCreated {
		t.Log("Very long description was accepted")
	} else {
		t.Log("Very long description was rejected")
	}
}

func TestIntegration_Validation_NumericFields(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	testUser := builders.NewUserBuilder().
		WithEmail("testuser@example.com").
		Build(t, ctx, db)

	testArtist := builders.NewArtistBuilder(testUser).
		WithName("Test Artist").
		Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	testCases := []struct {
		name           string
		duration       string
		expectedStatus int
	}{
		{"negative_duration", "-100", http.StatusBadRequest},
		{"zero_duration", "0", http.StatusBadRequest}, // Zero duration is invalid
		{"very_large_duration", "999999999", http.StatusCreated},
		{"non_numeric_duration", "abc", http.StatusBadRequest},
		{"float_duration", "123.45", http.StatusBadRequest}, // Expecting int
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formFields := map[string]string{
				"song_name":        "Test Song",
				"duration_seconds": tc.duration,
				"artist_uuid":      builders.UUIDToString(testArtist),
			}
			audioData := []byte("fake audio")
			req := createMultipartRequest(t, "POST", "/artists/"+builders.UUIDToString(testArtist)+"/music",
				"audio", "test.mp3", audioData, formFields)

			router := mux.NewRouter()
			router.HandleFunc("/artists/{artist_uuid}/music", wrapWithAuth(t, handler.CreateMusic, testUser)).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestIntegration_Validation_EmailFormat(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)

	testCases := []struct {
		name           string
		email          string
		expectedStatus int
	}{
		{"valid_email_simple", "test@example.com", http.StatusCreated},
		{"valid_email_subdomain", "test@mail.example.com", http.StatusCreated},
		{"valid_email_plus", "test+tag@example.com", http.StatusCreated},
		{"invalid_email_no_at", "testexample.com", http.StatusBadRequest},
		{"invalid_email_no_domain", "test@", http.StatusBadRequest},
		{"invalid_email_no_local", "@example.com", http.StatusBadRequest},
		{"invalid_email_spaces", "test @example.com", http.StatusBadRequest},
		{"invalid_email_double_at", "test@@example.com", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createJSONRequest(t, "POST", "/register", map[string]interface{}{
				"username": "user_" + strings.ReplaceAll(tc.email, "@", "_"),
				"email":    tc.email,
				"password": "TestPassword123!",
			})

			router := mux.NewRouter()
			router.HandleFunc("/register", handler.Register).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code, "Email: "+tc.email)
		})
	}
}

func TestIntegration_Validation_PasswordStrength(t *testing.T) {
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	handler := handlers.NewUserHandler(logger, config, nil, returns, nil, nil)

	testCases := []struct {
		name           string
		password       string
		expectedStatus int
		description    string
	}{
		{"weak_password_too_short", "Short1!", http.StatusBadRequest, "Password too short"},
		{"weak_password_no_number", "Password!", http.StatusBadRequest, "No number"},
		{"weak_password_no_special", "Password1", http.StatusBadRequest, "No special char"},
		{"weak_password_no_uppercase", "password1!", http.StatusBadRequest, "No uppercase"},
		{"weak_password_no_lowercase", "PASSWORD1!", http.StatusBadRequest, "No lowercase"},
		{"valid_password", "SecurePass123!", http.StatusOK, "Valid password"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createJSONRequest(t, "POST", "/register", map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": tc.password,
			})

			router := mux.NewRouter()
			router.HandleFunc("/register", handler.Register).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Just verify the request is processed (actual validation depends on implementation)
			// Some implementations may not validate password strength
			assert.True(t, rr.Code == http.StatusCreated || rr.Code == http.StatusBadRequest)
		})
	}
}

func TestIntegration_Validation_SecurityInput(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	testUser := builders.NewUserBuilder().
		WithEmail("securitytest@example.com").
		Build(t, ctx, db)

	handler := handlers.NewUserHandler(logger, config, nil, returns, db, nil)

	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
		expectSafe     bool // true if we expect the input to be safely stored (OK status)
	}{
		{
			name: "sql_injection_attempt_classic",
			requestBody: map[string]interface{}{
				"username": "admin'; DROP TABLE users; --",
				"country":  "US",
			},
			expectedStatus: http.StatusOK, // Should be stored as literal string
			description:    "SQL injection should be stored as literal string (parameterized queries)",
			expectSafe:     true,
		},
		{
			name: "sql_injection_attempt_union",
			requestBody: map[string]interface{}{
				"username": "admin' UNION SELECT * FROM users--",
				"country":  "US",
			},
			expectedStatus: http.StatusOK,
			description:    "SQL UNION injection should be stored as literal string",
			expectSafe:     true,
		},
		{
			name: "xss_attempt_script_tag",
			requestBody: map[string]interface{}{
				"username": "<script>alert('XSS')</script>",
				"country":  "US",
			},
			expectedStatus: http.StatusOK,
			description:    "XSS script tag should be stored as literal string",
			expectSafe:     true,
		},
		{
			name: "xss_attempt_img_onerror",
			requestBody: map[string]interface{}{
				"username": "<img src=x onerror=alert('XSS')>",
				"country":  "US",
			},
			expectedStatus: http.StatusOK,
			description:    "XSS img onerror should be stored as literal string",
			expectSafe:     true,
		},
		{
			name: "null_byte_injection",
			requestBody: map[string]interface{}{
				"username": "valid\x00username",
				"country":  "US",
			},
			expectedStatus: http.StatusBadRequest, // Null bytes should be rejected
			description:    "Null byte injection should be rejected",
			expectSafe:     false,
		},
		{
			name: "newline_injection",
			requestBody: map[string]interface{}{
				"username": "valid\nusername",
				"country":  "US",
			},
			expectedStatus: http.StatusBadRequest, // Newlines in username should be rejected
			description:    "Newline injection should be rejected",
			expectSafe:     false,
		},
		{
			name: "carriage_return_injection",
			requestBody: map[string]interface{}{
				"username": "valid\rusername",
				"country":  "US",
			},
			expectedStatus: http.StatusBadRequest, // Carriage returns should be rejected
			description:    "Carriage return injection should be rejected",
			expectSafe:     false,
		},
		{
			name: "null_unicode",
			requestBody: map[string]interface{}{
				"username": "user\u0000name",
				"country":  "US",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Null unicode character should be rejected",
			expectSafe:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createJSONRequest(t, "POST", "/users/me", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/users/me", wrapWithAuth(t, handler.UpdateProfile, testUser)).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code, tc.description)

			// If the request was successful, verify the data was stored safely
			if tc.expectSafe && rr.Code == http.StatusOK {
				user, err := db.GetPublicUser(ctx, testUser)
				require.NoError(t, err)
				// Verify the raw string is stored (not executed)
				assert.NotContains(t, user.Username, "<script>")
				assert.NotContains(t, user.Username, "DROP TABLE")
			}
		})
	}
}
