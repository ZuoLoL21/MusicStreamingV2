package handlers

import (
	libsclients "libs/clients"
	"libs/metrics"
	"net/http"
	"time"

	libshandlers "libs/handlers"
)

// ProxyHandler handles proxying requests to backend services
type ProxyHandler struct {
	userDBClient         *libsclients.ProxyClient
	recommendClient      *libsclients.ProxyClient
	eventIngestionClient *libsclients.ProxyClient
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(
	userDBClient *libsclients.ProxyClient,
	recommendClient *libsclients.ProxyClient,
	eventIngestionClient *libsclients.ProxyClient,
) *ProxyHandler {
	return &ProxyHandler{
		userDBClient:         userDBClient,
		recommendClient:      recommendClient,
		eventIngestionClient: eventIngestionClient,
	}
}

// ProxyLogin handles login requests (POST and PUT /login)
// These are public routes that don't require authentication
func (h *ProxyHandler) ProxyLogin(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyPublic(w, r, h.userDBClient.ForwardRequest)
}

// ProxyPublicFiles handles public file requests (GET /files/public/*)
func (h *ProxyHandler) ProxyPublicFiles(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyPublic(w, r, h.userDBClient.ForwardRequest)
}

// ProxyPrivateFiles handles public file requests (GET /files/private/*)
// Requires normal JWT validation (handled by middleware)
func (h *ProxyHandler) ProxyPrivateFiles(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.userDBClient.ForwardWithServiceJWT)
}

// ProxyRenew handles token renewal requests (POST /renew)
// Requires refresh token validation (handled by middleware)
// Forwards both service JWT (for auth) and original refresh token (for DB validation)
func (h *ProxyHandler) ProxyRenew(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyRenewWithServiceJWT(w, r, h.userDBClient.ForwardWithServiceJWT)
}

// ProxyUserDatabase handles all user database service routes
// Requires normal JWT validation (handled by middleware)
func (h *ProxyHandler) ProxyUserDatabase(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	libshandlers.ProxyWithServiceJWT(w, r, h.userDBClient.ForwardWithServiceJWT)
	metrics.TrackDownstreamCall("user_database", "/proxy", time.Since(start), nil)
}

// ProxyRecommendation handles all recommendation service routes
// Requires normal JWT validation (handled by middleware)
func (h *ProxyHandler) ProxyRecommendation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	libshandlers.ProxyWithServiceJWT(w, r, h.recommendClient.ForwardWithServiceJWT)
	metrics.TrackDownstreamCall("recommendation", "/proxy", time.Since(start), nil)
}

// ProxyEventIngestion handles all event ingestion service routes
// Requires normal JWT validation (handled by middleware)
func (h *ProxyHandler) ProxyEventIngestion(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	libshandlers.ProxyWithServiceJWT(w, r, h.eventIngestionClient.ForwardWithServiceJWT)
	metrics.TrackDownstreamCall("event_ingestion", "/proxy", time.Since(start), nil)
}
