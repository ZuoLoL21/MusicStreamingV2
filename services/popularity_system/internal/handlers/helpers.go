package handlers

import (
	"net/http"
	"strconv"
)

// parsePaginationDecay parses limit, cursor_decay (float64), and cursor_id
// from the request's URL query parameters. Used for decay-based popularity
// queries (all-time rankings).
//
// Returns:
//   - limit: default 50, max 100
//   - cursorDecay: 0.0 if not provided (indicates first page)
//   - cursorID: empty string if not provided
func parsePaginationDecay(r *http.Request) (limit int, cursorDecay float64, cursorID string) {
	limit = parseLimit(r)

	if s := r.URL.Query().Get("cursor_decay"); s != "" {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			cursorDecay = f
		}
	}

	cursorID = r.URL.Query().Get("cursor_id")
	return
}

// parsePaginationPlays parses limit, cursor_plays (uint64), and cursor_id
// from the request's URL query parameters. Used for play-count-based queries
// (timeframe rankings).
//
// Returns:
//   - limit: default 50, max 100
//   - cursorPlays: 0 if not provided (indicates first page)
//   - cursorID: empty string if not provided
func parsePaginationPlays(r *http.Request) (limit int, cursorPlays uint64, cursorID string) {
	limit = parseLimit(r)

	if s := r.URL.Query().Get("cursor_plays"); s != "" {
		if u, err := strconv.ParseUint(s, 10, 64); err == nil {
			cursorPlays = u
		}
	}

	cursorID = r.URL.Query().Get("cursor_id")
	return
}

// parsePaginationTheme parses limit, cursor_decay, and cursor_theme
// from the request's URL query parameters. Used for theme popularity rankings.
//
// Returns:
//   - limit: default 50, max 100
//   - cursorDecay: 0.0 if not provided (indicates first page)
//   - cursorTheme: empty string if not provided
func parsePaginationTheme(r *http.Request) (limit int, cursorDecay float64, cursorTheme string) {
	limit = parseLimit(r)

	if s := r.URL.Query().Get("cursor_decay"); s != "" {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			cursorDecay = f
		}
	}

	cursorTheme = r.URL.Query().Get("cursor_theme")
	return
}
