package handlers

import (
	"gateway_recommendation/internal/clients"
	"net/http"

	libshandlers "libs/handlers"
)

type ProxyHandler struct {
	popularityClient *clients.PopularityClient
}

func NewProxyHandler(popularityClient *clients.PopularityClient) *ProxyHandler {
	return &ProxyHandler{
		popularityClient: popularityClient,
	}
}

// ProxyPopularSongsAllTime handles GET /popular/songs/all-time
func (h *ProxyHandler) ProxyPopularSongsAllTime(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.popularityClient.ForwardWithServiceJWT)
}

// ProxyPopularArtistsAllTime handles GET /popular/artists/all-time
func (h *ProxyHandler) ProxyPopularArtistsAllTime(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.popularityClient.ForwardWithServiceJWT)
}

// ProxyPopularThemesAllTime handles GET /popular/themes/all-time
func (h *ProxyHandler) ProxyPopularThemesAllTime(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.popularityClient.ForwardWithServiceJWT)
}

// ProxyPopularSongsByTheme handles GET /popular/songs/theme/{theme}
func (h *ProxyHandler) ProxyPopularSongsByTheme(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.popularityClient.ForwardWithServiceJWT)
}

// ProxyPopularSongsTimeframe handles GET /popular/songs/timeframe
func (h *ProxyHandler) ProxyPopularSongsTimeframe(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.popularityClient.ForwardWithServiceJWT)
}

// ProxyPopularArtistsTimeframe handles GET /popular/artists/timeframe
func (h *ProxyHandler) ProxyPopularArtistsTimeframe(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.popularityClient.ForwardWithServiceJWT)
}

// ProxyPopularThemesTimeframe handles GET /popular/themes/timeframe
func (h *ProxyHandler) ProxyPopularThemesTimeframe(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.popularityClient.ForwardWithServiceJWT)
}

// ProxyPopularSongsByThemeTimeframe handles GET /popular/songs/theme/{theme}/timeframe
func (h *ProxyHandler) ProxyPopularSongsByThemeTimeframe(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.popularityClient.ForwardWithServiceJWT)
}
