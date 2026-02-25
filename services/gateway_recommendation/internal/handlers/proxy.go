package handlers

import (
	"context"
	"gateway_recommendation/internal/clients"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type ProxyHandler struct {
	popularityClient *clients.PopularityClient
	logger           *zap.Logger
	requestIDKey     any
}

func NewProxyHandler(popularityClient *clients.PopularityClient, logger *zap.Logger, requestIDKey any) *ProxyHandler {
	return &ProxyHandler{
		popularityClient: popularityClient,
		logger:           logger,
		requestIDKey:     requestIDKey,
	}
}

func (h *ProxyHandler) getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(h.requestIDKey).(string); ok {
		return reqID
	}
	return "unknown"
}

// proxyToPopularity is a helper that forwards requests to the popularity service
func (h *ProxyHandler) proxyToPopularity(w http.ResponseWriter, r *http.Request, path string) {
	requestID := h.getRequestID(r.Context())

	body, statusCode, err := h.popularityClient.ProxyRequest(
		r.Context(),
		r.Method,
		path,
		r.URL.RawQuery,
		requestID,
	)

	if err != nil {
		h.logger.Error("Proxy request failed",
			zap.String("request_id", requestID),
			zap.String("path", path),
			zap.Error(err))
		http.Error(w, "Failed to proxy request", statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

// ProxyPopularSongsAllTime handles GET /popular/songs/all-time
func (h *ProxyHandler) ProxyPopularSongsAllTime(w http.ResponseWriter, r *http.Request) {
	h.proxyToPopularity(w, r, "/popular/songs/all-time")
}

// ProxyPopularArtistsAllTime handles GET /popular/artists/all-time
func (h *ProxyHandler) ProxyPopularArtistsAllTime(w http.ResponseWriter, r *http.Request) {
	h.proxyToPopularity(w, r, "/popular/artists/all-time")
}

// ProxyPopularThemesAllTime handles GET /popular/themes/all-time
func (h *ProxyHandler) ProxyPopularThemesAllTime(w http.ResponseWriter, r *http.Request) {
	h.proxyToPopularity(w, r, "/popular/themes/all-time")
}

// ProxyPopularSongsByTheme handles GET /popular/songs/theme/{theme}
func (h *ProxyHandler) ProxyPopularSongsByTheme(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	theme := vars["theme"]
	path := "/popular/songs/theme/" + theme
	h.proxyToPopularity(w, r, path)
}

// ProxyPopularSongsTimeframe handles GET /popular/songs/timeframe
func (h *ProxyHandler) ProxyPopularSongsTimeframe(w http.ResponseWriter, r *http.Request) {
	h.proxyToPopularity(w, r, "/popular/songs/timeframe")
}

// ProxyPopularArtistsTimeframe handles GET /popular/artists/timeframe
func (h *ProxyHandler) ProxyPopularArtistsTimeframe(w http.ResponseWriter, r *http.Request) {
	h.proxyToPopularity(w, r, "/popular/artists/timeframe")
}

// ProxyPopularThemesTimeframe handles GET /popular/themes/timeframe
func (h *ProxyHandler) ProxyPopularThemesTimeframe(w http.ResponseWriter, r *http.Request) {
	h.proxyToPopularity(w, r, "/popular/themes/timeframe")
}

// ProxyPopularSongsByThemeTimeframe handles GET /popular/songs/theme/{theme}/timeframe
func (h *ProxyHandler) ProxyPopularSongsByThemeTimeframe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	theme := vars["theme"]
	path := "/popular/songs/theme/" + theme + "/timeframe"
	h.proxyToPopularity(w, r, path)
}
