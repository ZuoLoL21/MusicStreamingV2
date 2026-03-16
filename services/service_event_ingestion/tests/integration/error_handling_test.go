//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_Error_InvalidJSON(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name     string
		endpoint string
		handler  func(http.ResponseWriter, *http.Request)
	}{
		{
			name:     "invalid_json_listen",
			endpoint: "/api/v1/events/listen",
			handler:  handler.IngestListenEvent,
		},
		{
			name:     "invalid_json_like",
			endpoint: "/api/v1/events/like",
			handler:  handler.IngestLikeEvent,
		},
		{
			name:     "invalid_json_theme",
			endpoint: "/api/v1/events/theme",
			handler:  handler.IngestThemeEvent,
		},
		{
			name:     "invalid_json_user",
			endpoint: "/api/v1/events/user",
			handler:  handler.IngestUserDimEvent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request with invalid JSON body
			req := httptest.NewRequest("POST", tc.endpoint, strings.NewReader("{invalid json"))
			req.Header.Set("Content-Type", "application/json")

			router := mux.NewRouter()
			router.HandleFunc(tc.endpoint, tc.handler).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestIntegration_Error_MalformedJSON(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name     string
		endpoint string
		body     string
		handler  func(http.ResponseWriter, *http.Request)
	}{
		{
			name:     "malformed_json_listen",
			endpoint: "/api/v1/events/listen",
			body:     `{"user_uuid": "not-a-uuid"}`,
			handler:  handler.IngestListenEvent,
		},
		{
			name:     "malformed_json_like",
			endpoint: "/api/v1/events/like",
			body:     `{"user_uuid": "not-a-uuid"}`,
			handler:  handler.IngestLikeEvent,
		},
		{
			name:     "malformed_json_theme",
			endpoint: "/api/v1/events/theme",
			body:     `{"music_uuid": "not-a-uuid"}`,
			handler:  handler.IngestThemeEvent,
		},
		{
			name:     "malformed_json_user",
			endpoint: "/api/v1/events/user",
			body:     `{"country": "USA"}`,
			handler:  handler.IngestUserDimEvent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tc.endpoint, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")

			router := mux.NewRouter()
			router.HandleFunc(tc.endpoint, tc.handler).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Should return 400 Bad Request for malformed/invalid UUIDs
			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestIntegration_Error_MethodNotAllowed(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name     string
		endpoint string
		handler  func(http.ResponseWriter, *http.Request)
	}{
		{
			name:     "get_on_listen_endpoint",
			endpoint: "/api/v1/events/listen",
			handler:  handler.IngestListenEvent,
		},
		{
			name:     "get_on_like_endpoint",
			endpoint: "/api/v1/events/like",
			handler:  handler.IngestLikeEvent,
		},
		{
			name:     "get_on_theme_endpoint",
			endpoint: "/api/v1/events/theme",
			handler:  handler.IngestThemeEvent,
		},
		{
			name:     "get_on_user_endpoint",
			endpoint: "/api/v1/events/user",
			handler:  handler.IngestUserDimEvent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Try GET on POST endpoint
			req := httptest.NewRequest("GET", tc.endpoint, nil)

			router := mux.NewRouter()
			router.HandleFunc(tc.endpoint, tc.handler).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
		})
	}
}

func TestIntegration_Error_EmptyBody(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name     string
		endpoint string
		handler  func(http.ResponseWriter, *http.Request)
	}{
		{
			name:     "empty_body_listen",
			endpoint: "/api/v1/events/listen",
			handler:  handler.IngestListenEvent,
		},
		{
			name:     "empty_body_like",
			endpoint: "/api/v1/events/like",
			handler:  handler.IngestLikeEvent,
		},
		{
			name:     "empty_body_theme",
			endpoint: "/api/v1/events/theme",
			handler:  handler.IngestThemeEvent,
		},
		{
			name:     "empty_body_user",
			endpoint: "/api/v1/events/user",
			handler:  handler.IngestUserDimEvent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tc.endpoint, strings.NewReader(""))
			req.Header.Set("Content-Type", "application/json")

			router := mux.NewRouter()
			router.HandleFunc(tc.endpoint, tc.handler).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Should return 400 Bad Request for empty body
			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestIntegration_Error_WrongContentType(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name     string
		endpoint string
		handler  func(http.ResponseWriter, *http.Request)
	}{
		{
			name:     "text_plain_listen",
			endpoint: "/api/v1/events/listen",
			handler:  handler.IngestListenEvent,
		},
		{
			name:     "text_plain_like",
			endpoint: "/api/v1/events/like",
			handler:  handler.IngestLikeEvent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tc.endpoint, strings.NewReader("plain text"))
			req.Header.Set("Content-Type", "text/plain")

			router := mux.NewRouter()
			router.HandleFunc(tc.endpoint, tc.handler).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Should return 400 or 415 for wrong content type
			assert.True(t, rr.Code == http.StatusBadRequest || rr.Code == http.StatusUnsupportedMediaType)
		})
	}
}
