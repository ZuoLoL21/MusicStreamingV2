//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_Validation_ListenEvent(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name: "valid_listen_event",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 120,
				"track_duration_seconds":  180,
				"completion_ratio":        0.67,
			},
			expectedStatus: http.StatusOK,
			description:    "Valid listen event should be accepted",
		},
		{
			name: "missing_user_uuid",
			requestBody: map[string]interface{}{
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 120,
				"track_duration_seconds":  180,
				"completion_ratio":        0.67,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Missing user_uuid should be rejected",
		},
		{
			name: "missing_music_uuid",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 120,
				"track_duration_seconds":  180,
				"completion_ratio":        0.67,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Missing music_uuid should be rejected",
		},
		{
			name: "missing_artist_uuid",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"listen_duration_seconds": 120,
				"track_duration_seconds":  180,
				"completion_ratio":        0.67,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Missing artist_uuid should be rejected",
		},
		{
			name: "completion_ratio_negative",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 120,
				"track_duration_seconds":  180,
				"completion_ratio":        -0.5,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Negative completion_ratio should be rejected",
		},
		{
			name: "completion_ratio_over_1",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 120,
				"track_duration_seconds":  180,
				"completion_ratio":        1.5,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "completion_ratio > 1 should be rejected",
		},
		{
			name: "completion_ratio_exactly_1",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 180,
				"track_duration_seconds":  180,
				"completion_ratio":        1.0,
			},
			expectedStatus: http.StatusOK,
			description:    "completion_ratio exactly 1.0 should be accepted",
		},
		{
			name: "completion_ratio_zero",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 0,
				"track_duration_seconds":  180,
				"completion_ratio":        0.0,
			},
			expectedStatus: http.StatusOK,
			description:    "completion_ratio exactly 0.0 should be accepted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createJSONRequest(t, "POST", "/api/v1/events/listen", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/api/v1/events/listen", handler.IngestListenEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code, tc.description)
		})
	}
}

func TestIntegration_Validation_LikeEvent(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name: "valid_like_event",
			requestBody: map[string]interface{}{
				"user_uuid":   NewTestUUID(),
				"music_uuid":  NewTestUUID(),
				"artist_uuid": NewTestUUID(),
			},
			expectedStatus: http.StatusOK,
			description:    "Valid like event should be accepted",
		},
		{
			name: "missing_user_uuid",
			requestBody: map[string]interface{}{
				"music_uuid":  NewTestUUID(),
				"artist_uuid": NewTestUUID(),
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Missing user_uuid should be rejected",
		},
		{
			name: "missing_music_uuid",
			requestBody: map[string]interface{}{
				"user_uuid":   NewTestUUID(),
				"artist_uuid": NewTestUUID(),
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Missing music_uuid should be rejected",
		},
		{
			name: "missing_artist_uuid",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"music_uuid": NewTestUUID(),
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Missing artist_uuid should be rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createJSONRequest(t, "POST", "/api/v1/events/like", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/api/v1/events/like", handler.IngestLikeEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code, tc.description)
		})
	}
}

func TestIntegration_Validation_ThemeEvent(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name: "valid_theme_event",
			requestBody: map[string]interface{}{
				"music_uuid": NewTestUUID(),
				"theme":      "rock",
			},
			expectedStatus: http.StatusOK,
			description:    "Valid theme event should be accepted",
		},
		{
			name: "missing_music_uuid",
			requestBody: map[string]interface{}{
				"theme": "rock",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Missing music_uuid should be rejected",
		},
		{
			name: "missing_theme",
			requestBody: map[string]interface{}{
				"music_uuid": NewTestUUID(),
			},
			expectedStatus: http.StatusBadRequest,
			description:    " Missing theme should be rejected",
		},
		{
			name: "empty_theme",
			requestBody: map[string]interface{}{
				"music_uuid": NewTestUUID(),
				"theme":      "",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Empty theme should be rejected",
		},
		{
			name: "theme_with_special_chars",
			requestBody: map[string]interface{}{
				"music_uuid": NewTestUUID(),
				"theme":      "rock-n-roll",
			},
			expectedStatus: http.StatusOK,
			description:    "Theme with special chars should be accepted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createJSONRequest(t, "POST", "/api/v1/events/theme", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/api/v1/events/theme", handler.IngestThemeEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code, tc.description)
		})
	}
}

func TestIntegration_Validation_UserDimEvent(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name: "valid_user_dim_event",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"country":    "US",
				"created_at": "2024-01-01T00:00:00Z",
			},
			expectedStatus: http.StatusOK,
			description:    "Valid user dimension event should be accepted",
		},
		{
			name: "missing_user_uuid",
			requestBody: map[string]interface{}{
				"country":    "US",
				"created_at": "2024-01-01T00:00:00Z",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Missing user_uuid should be rejected",
		},
		{
			name: "missing_country",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"created_at": "2024-01-01T00:00:00Z",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Missing country should be rejected",
		},
		{
			name: "country_too_long",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"country":    "USA",
				"created_at": "2024-01-01T00:00:00Z",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Country code > 2 chars should be rejected",
		},
		{
			name: "country_one_char",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"country":    "U",
				"created_at": "2024-01-01T00:00:00Z",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Country code < 2 chars should be rejected",
		},
		{
			name: "valid_two_char_country",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"country":    "CA",
				"created_at": "2024-01-01T00:00:00Z",
			},
			expectedStatus: http.StatusOK,
			description:    "Valid 2-char country code should be accepted",
		},
		{
			name: "lowercase_country",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"country":    "us",
				"created_at": "2024-01-01T00:00:00Z",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Lowercase country code should be rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createJSONRequest(t, "POST", "/api/v1/events/user", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/api/v1/events/user", handler.IngestUserDimEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code, tc.description)
		})
	}
}
