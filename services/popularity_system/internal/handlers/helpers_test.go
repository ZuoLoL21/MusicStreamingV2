package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePaginationDecay(t *testing.T) {
	tests := []struct {
		name                string
		url                 string
		expectedLimit       int
		expectedCursorDecay float64
		expectedCursorID    string
	}{
		{
			name:                "no parameters - defaults",
			url:                 "/test",
			expectedLimit:       50,
			expectedCursorDecay: 0.0,
			expectedCursorID:    "",
		},
		{
			name:                "with valid limit",
			url:                 "/test?limit=25",
			expectedLimit:       25,
			expectedCursorDecay: 0.0,
			expectedCursorID:    "",
		},
		{
			name:                "limit exceeds max - capped at 100",
			url:                 "/test?limit=200",
			expectedLimit:       50, // invalid, defaults to 50
			expectedCursorDecay: 0.0,
			expectedCursorID:    "",
		},
		{
			name:                "with cursor values",
			url:                 "/test?limit=10&cursor_decay=1234.56&cursor_id=abc-123",
			expectedLimit:       10,
			expectedCursorDecay: 1234.56,
			expectedCursorID:    "abc-123",
		},
		{
			name:                "invalid decay - uses default 0.0",
			url:                 "/test?cursor_decay=invalid",
			expectedLimit:       50,
			expectedCursorDecay: 0.0,
			expectedCursorID:    "",
		},
		{
			name:                "zero limit - uses default",
			url:                 "/test?limit=0",
			expectedLimit:       50,
			expectedCursorDecay: 0.0,
			expectedCursorID:    "",
		},
		{
			name:                "negative limit - uses default",
			url:                 "/test?limit=-10",
			expectedLimit:       50,
			expectedCursorDecay: 0.0,
			expectedCursorID:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			limit, cursorDecay, cursorID := parsePaginationDecay(req)

			if limit != tt.expectedLimit {
				t.Errorf("limit: got %d, want %d", limit, tt.expectedLimit)
			}
			if cursorDecay != tt.expectedCursorDecay {
				t.Errorf("cursorDecay: got %f, want %f", cursorDecay, tt.expectedCursorDecay)
			}
			if cursorID != tt.expectedCursorID {
				t.Errorf("cursorID: got %q, want %q", cursorID, tt.expectedCursorID)
			}
		})
	}
}

func TestParsePaginationPlays(t *testing.T) {
	tests := []struct {
		name                string
		url                 string
		expectedLimit       int
		expectedCursorPlays uint64
		expectedCursorID    string
	}{
		{
			name:                "no parameters - defaults",
			url:                 "/test",
			expectedLimit:       50,
			expectedCursorPlays: 0,
			expectedCursorID:    "",
		},
		{
			name:                "with valid limit",
			url:                 "/test?limit=30",
			expectedLimit:       30,
			expectedCursorPlays: 0,
			expectedCursorID:    "",
		},
		{
			name:                "with cursor values",
			url:                 "/test?limit=20&cursor_plays=5000&cursor_id=xyz-789",
			expectedLimit:       20,
			expectedCursorPlays: 5000,
			expectedCursorID:    "xyz-789",
		},
		{
			name:                "invalid plays - uses default 0",
			url:                 "/test?cursor_plays=invalid",
			expectedLimit:       50,
			expectedCursorPlays: 0,
			expectedCursorID:    "",
		},
		{
			name:                "negative plays - uses default 0",
			url:                 "/test?cursor_plays=-100",
			expectedLimit:       50,
			expectedCursorPlays: 0,
			expectedCursorID:    "",
		},
		{
			name:                "large plays value",
			url:                 "/test?cursor_plays=9999999999",
			expectedLimit:       50,
			expectedCursorPlays: 9999999999,
			expectedCursorID:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			limit, cursorPlays, cursorID := parsePaginationPlays(req)

			if limit != tt.expectedLimit {
				t.Errorf("limit: got %d, want %d", limit, tt.expectedLimit)
			}
			if cursorPlays != tt.expectedCursorPlays {
				t.Errorf("cursorPlays: got %d, want %d", cursorPlays, tt.expectedCursorPlays)
			}
			if cursorID != tt.expectedCursorID {
				t.Errorf("cursorID: got %q, want %q", cursorID, tt.expectedCursorID)
			}
		})
	}
}

func TestParsePaginationTheme(t *testing.T) {
	tests := []struct {
		name                string
		url                 string
		expectedLimit       int
		expectedCursorDecay float64
		expectedCursorTheme string
	}{
		{
			name:                "no parameters - defaults",
			url:                 "/test",
			expectedLimit:       50,
			expectedCursorDecay: 0.0,
			expectedCursorTheme: "",
		},
		{
			name:                "with valid limit",
			url:                 "/test?limit=15",
			expectedLimit:       15,
			expectedCursorDecay: 0.0,
			expectedCursorTheme: "",
		},
		{
			name:                "with cursor values",
			url:                 "/test?limit=10&cursor_decay=888.99&cursor_theme=rock",
			expectedLimit:       10,
			expectedCursorDecay: 888.99,
			expectedCursorTheme: "rock",
		},
		{
			name:                "invalid decay - uses default 0.0",
			url:                 "/test?cursor_decay=notanumber&cursor_theme=jazz",
			expectedLimit:       50,
			expectedCursorDecay: 0.0,
			expectedCursorTheme: "jazz",
		},
		{
			name:                "theme with spaces",
			url:                 "/test?cursor_theme=classic%20rock",
			expectedLimit:       50,
			expectedCursorDecay: 0.0,
			expectedCursorTheme: "classic rock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			limit, cursorDecay, cursorTheme := parsePaginationTheme(req)

			if limit != tt.expectedLimit {
				t.Errorf("limit: got %d, want %d", limit, tt.expectedLimit)
			}
			if cursorDecay != tt.expectedCursorDecay {
				t.Errorf("cursorDecay: got %f, want %f", cursorDecay, tt.expectedCursorDecay)
			}
			if cursorTheme != tt.expectedCursorTheme {
				t.Errorf("cursorTheme: got %q, want %q", cursorTheme, tt.expectedCursorTheme)
			}
		})
	}
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedLimit int
	}{
		{
			name:          "no limit - default 50",
			url:           "/test",
			expectedLimit: 50,
		},
		{
			name:          "valid limit within range",
			url:           "/test?limit=75",
			expectedLimit: 75,
		},
		{
			name:          "limit exceeds max 100 - use default",
			url:           "/test?limit=150",
			expectedLimit: 50,
		},
		{
			name:          "limit at max 100",
			url:           "/test?limit=100",
			expectedLimit: 100,
		},
		{
			name:          "limit at min 1",
			url:           "/test?limit=1",
			expectedLimit: 1,
		},
		{
			name:          "zero limit - use default",
			url:           "/test?limit=0",
			expectedLimit: 50,
		},
		{
			name:          "negative limit - use default",
			url:           "/test?limit=-5",
			expectedLimit: 50,
		},
		{
			name:          "invalid limit - use default",
			url:           "/test?limit=abc",
			expectedLimit: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			limit := parseLimit(req)

			if limit != tt.expectedLimit {
				t.Errorf("got %d, want %d", limit, tt.expectedLimit)
			}
		})
	}
}
