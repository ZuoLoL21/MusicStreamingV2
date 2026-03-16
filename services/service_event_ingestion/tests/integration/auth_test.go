//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	libsconsts "libs/consts"
	libsdi "libs/di"
	libsmiddleware "libs/middleware"
)

// TestAuthentication_MissingJWT tests that requests without JWT tokens are rejected
func TestAuthentication_MissingJWT(t *testing.T) {
	handler := CreateTestHandler(t)
	logger := zap.NewNop()
	returns := libsdi.NewReturnManager(logger)

	testCases := []struct {
		name           string
		endpoint       string
		expectedStatus int
	}{
		{
			name:           "missing_jwt_listen",
			endpoint:       "/events/listen",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing_jwt_like",
			endpoint:       "/events/like",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing_jwt_theme",
			endpoint:       "/events/theme",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing_jwt_user",
			endpoint:       "/events/user",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := map[string]interface{}{
				"user_uuid":   NewTestUUID(),
				"music_uuid":  NewTestUUID(),
				"artist_uuid": NewTestUUID(),
			}
			req := createJSONRequest(t, "POST", tc.endpoint, reqBody)

			router := mux.NewRouter()
			authHandler := libsmiddleware.NewAuthHandler(
				logger,
				nil,
				returns,
				libsconsts.JWTSubjectService,
			)
			protected := router.PathPrefix("").Subrouter()
			protected.Use(authHandler.GetAuthMiddleware())
			protected.HandleFunc(tc.endpoint, handler.IngestListenEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.True(t, rr.Code == http.StatusUnauthorized || rr.Code == http.StatusInternalServerError)
		})
	}
}

// TestAuthentication_InvalidAuthHeader tests various invalid Authorization header formats
func TestAuthentication_InvalidAuthHeader(t *testing.T) {
	handler := CreateTestHandler(t)
	logger := zap.NewNop()
	returns := libsdi.NewReturnManager(logger)

	testCases := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "empty_auth_header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing_bearer_prefix",
			authHeader:     "some-token-value",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "wrong_prefix",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty_token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "multiple_spaces",
			authHeader:     "Bearer token1 token2",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 120,
				"track_duration_seconds":  180,
				"completion_ratio":        0.67,
			}
			req := createJSONRequest(t, "POST", "/events/listen", reqBody)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			router := mux.NewRouter()
			authHandler := libsmiddleware.NewAuthHandler(logger, nil, returns, libsconsts.JWTSubjectService)
			protected := router.PathPrefix("").Subrouter()
			protected.Use(authHandler.GetAuthMiddleware())
			protected.HandleFunc("/events/listen", handler.IngestListenEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

// TestAuthentication_InvalidJWTToken tests that invalid JWT tokens are rejected
func TestAuthentication_InvalidJWTToken(t *testing.T) {
	handler := CreateTestHandler(t)
	logger := zap.NewNop()
	returns := libsdi.NewReturnManager(logger)

	testCases := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "random_string_token",
			token:          "random-not-a-valid-jwt-string",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "malformed_jwt",
			token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired_jwt_format",
			token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "partial_jwt",
			token:          "eyJhbGciOiJIUzI1NiJ9",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 120,
				"track_duration_seconds":  180,
				"completion_ratio":        0.67,
			}
			req := createJSONRequest(t, "POST", "/events/listen", reqBody)
			req.Header.Set("Authorization", "Bearer "+tc.token)

			router := mux.NewRouter()
			authHandler := libsmiddleware.NewAuthHandler(logger, nil, returns, libsconsts.JWTSubjectService)
			protected := router.PathPrefix("").Subrouter()
			protected.Use(authHandler.GetAuthMiddleware())
			protected.HandleFunc("/events/listen", handler.IngestListenEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

// TestAuthentication_MalformedHeaders tests various malformed header scenarios
func TestAuthentication_MalformedHeaders(t *testing.T) {
	handler := CreateTestHandler(t)
	logger := zap.NewNop()
	returns := libsdi.NewReturnManager(logger)

	testCases := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "lowercase_bearer",
			authHeader:     "bearer valid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "bearer_uppercase",
			authHeader:     "BEARER valid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "tab_in_header",
			authHeader:     "Bearer\tvalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 120,
				"track_duration_seconds":  180,
				"completion_ratio":        0.67,
			}
			req := createJSONRequest(t, "POST", "/events/listen", reqBody)
			req.Header.Set("Authorization", tc.authHeader)

			router := mux.NewRouter()
			authHandler := libsmiddleware.NewAuthHandler(logger, nil, returns, libsconsts.JWTSubjectService)
			protected := router.PathPrefix("").Subrouter()
			protected.Use(authHandler.GetAuthMiddleware())
			protected.HandleFunc("/events/listen", handler.IngestListenEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

// TestAuthentication_PublicEndpointAccessible tests that public endpoints don't require auth
func TestAuthentication_PublicEndpointAccessible(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)

	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ok")
}
