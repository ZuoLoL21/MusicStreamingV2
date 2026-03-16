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
			req := createJSONRequest(t, "POST", "/events/listen", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/events/listen", handler.IngestListenEvent).Methods("POST")

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
			req := createJSONRequest(t, "POST", "/events/like", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/events/like", handler.IngestLikeEvent).Methods("POST")

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
			req := createJSONRequest(t, "POST", "/events/theme", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/events/theme", handler.IngestThemeEvent).Methods("POST")

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
		{
			name: "invalid_country_numbers",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"country":    "12",
				"created_at": "2024-01-01T00:00:00Z",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Numeric country code should be rejected",
		},
		{
			name: "whitespace_country",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"country":    "US ",
				"created_at": "2024-01-01T00:00:00Z",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Country with trailing whitespace should be rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createJSONRequest(t, "POST", "/events/user", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/events/user", handler.IngestUserDimEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code, tc.description)
		})
	}
}

// TestIntegration_Validation_ListenEventEdgeCases tests additional edge cases for listen events
func TestIntegration_Validation_ListenEventEdgeCases(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name: "listen_duration_greater_than_track",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 300,
				"track_duration_seconds":  180,
				"completion_ratio":        1.0,
			},
			expectedStatus: http.StatusOK,
			description:    "Listen duration > track duration should be accepted (user skipped ahead)",
		},
		{
			name: "zero_duration_valid",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 0,
				"track_duration_seconds":  180,
				"completion_ratio":        0.0,
			},
			expectedStatus: http.StatusOK,
			description:    "Zero duration should be accepted (instant skip)",
		},
		{
			name: "completion_ratio_boundary_0",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 0,
				"track_duration_seconds":  180,
				"completion_ratio":        0.0,
			},
			expectedStatus: http.StatusOK,
			description:    "Completion ratio exactly 0 should be accepted",
		},
		{
			name: "completion_ratio_boundary_1",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 180,
				"track_duration_seconds":  180,
				"completion_ratio":        1.0,
			},
			expectedStatus: http.StatusOK,
			description:    "Completion ratio exactly 1 should be accepted",
		},
		{
			name: "completion_ratio_negative_rejected",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 60,
				"track_duration_seconds":  180,
				"completion_ratio":        -0.1,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Negative completion ratio should be rejected",
		},
		{
			name: "completion_ratio_over_1_rejected",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 200,
				"track_duration_seconds":  180,
				"completion_ratio":        1.1,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Completion ratio > 1 should be rejected",
		},
		{
			name: "optional_album_uuid_missing",
			requestBody: map[string]interface{}{
				"user_uuid":               NewTestUUID(),
				"music_uuid":              NewTestUUID(),
				"artist_uuid":             NewTestUUID(),
				"listen_duration_seconds": 120,
				"track_duration_seconds":  180,
				"completion_ratio":        0.67,
			},
			expectedStatus: http.StatusOK,
			description:    "Missing optional album_uuid should be accepted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createJSONRequest(t, "POST", "/events/listen", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/events/listen", handler.IngestListenEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code, tc.description)
		})
	}
}

// TestIntegration_Validation_ThemeEventEdgeCases tests additional edge cases for theme events
func TestIntegration_Validation_ThemeEventEdgeCases(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name: "theme_whitespace_only",
			requestBody: map[string]interface{}{
				"music_uuid": NewTestUUID(),
				"theme":      "   ",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Whitespace-only theme should be rejected",
		},
		{
			name: "theme_with_numbers",
			requestBody: map[string]interface{}{
				"music_uuid": NewTestUUID(),
				"theme":      "pop123",
			},
			expectedStatus: http.StatusOK,
			description:    "Theme with numbers should be accepted",
		},
		{
			name: "theme_unicode_chars",
			requestBody: map[string]interface{}{
				"music_uuid": NewTestUUID(),
				"theme":      "électronique",
			},
			expectedStatus: http.StatusOK,
			description:    "Theme with unicode characters should be accepted",
		},
		{
			name: "theme_very_long",
			requestBody: map[string]interface{}{
				"music_uuid": NewTestUUID(),
				"theme":      "a",
			},
			expectedStatus: http.StatusOK,
			description:    "Very long theme should be accepted",
		},
		{
			name: "theme_with_hyphen",
			requestBody: map[string]interface{}{
				"music_uuid": NewTestUUID(),
				"theme":      "rock-n-roll",
			},
			expectedStatus: http.StatusOK,
			description:    "Theme with hyphen should be accepted",
		},
		{
			name: "theme_with_underscore",
			requestBody: map[string]interface{}{
				"music_uuid": NewTestUUID(),
				"theme":      "hip_hop",
			},
			expectedStatus: http.StatusOK,
			description:    "Theme with underscore should be accepted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createJSONRequest(t, "POST", "/events/theme", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/events/theme", handler.IngestThemeEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code, tc.description)
		})
	}
}

// TestIntegration_Validation_UserDimEventEdgeCases tests additional edge cases for user dim events
func TestIntegration_Validation_UserDimEventEdgeCases(t *testing.T) {
	handler := CreateTestHandler(t)

	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name: "user_dim_invalid_date_format",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"country":    "US",
				"created_at": "not-a-date",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Invalid date format should be rejected",
		},
		{
			name: "user_dim_future_date",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"country":    "US",
				"created_at": "2099-12-31T23:59:59Z",
			},
			expectedStatus: http.StatusOK,
			description:    "Future date should be accepted",
		},
		{
			name: "user_dim_past_date",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"country":    "US",
				"created_at": "1980-01-01T00:00:00Z",
			},
			expectedStatus: http.StatusOK,
			description:    "Past date should be accepted",
		},
		{
			name: "user_dim_missing_created_at",
			requestBody: map[string]interface{}{
				"user_uuid": NewTestUUID(),
				"country":   "US",
			},
			expectedStatus: http.StatusOK,
			description:    "Missing created_at should use current time",
		},
		{
			name: "user_dim_invalid_country_special_chars",
			requestBody: map[string]interface{}{
				"user_uuid":  NewTestUUID(),
				"country":    "U/S",
				"created_at": "2024-01-01T00:00:00Z",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Country with special chars should be rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createJSONRequest(t, "POST", "/events/user", tc.requestBody)

			router := mux.NewRouter()
			router.HandleFunc("/events/user", handler.IngestUserDimEvent).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code, tc.description)
		})
	}
}
