package handlers

import (
	"gateway_api/internal/clients"
	"net/http"

	libshandlers "libs/handlers"

	"go.uber.org/zap"
)

// ProxyHandler handles proxying requests to backend services
type ProxyHandler struct {
	userDBClient         *clients.UserDatabaseClient
	recommendClient      *clients.RecommendationClient
	eventIngestionClient *clients.EventIngestionClient
	logger               *zap.Logger
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(
	userDBClient *clients.UserDatabaseClient,
	recommendClient *clients.RecommendationClient,
	eventIngestionClient *clients.EventIngestionClient,
	logger *zap.Logger,
) *ProxyHandler {
	return &ProxyHandler{
		userDBClient:         userDBClient,
		recommendClient:      recommendClient,
		eventIngestionClient: eventIngestionClient,
		logger:               logger,
	}
}

// ProxyLogin handles login requests (POST and PUT /login)
// These are public routes that don't require authentication
func (h *ProxyHandler) ProxyLogin(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyPublic(w, r, h.logger, h.userDBClient.ForwardRequest)
}

// ProxyPublicFiles handles public file requests (GET /files/public/*)
func (h *ProxyHandler) ProxyPublicFiles(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyPublic(w, r, h.logger, h.userDBClient.ForwardRequest)
}

// ProxyPrivateFiles handles public file requests (GET /files/private/*)
// Requires normal JWT validation (handled by middleware)
func (h *ProxyHandler) ProxyPrivateFiles(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.logger, h.userDBClient.ForwardWithServiceJWT)
}

// ProxyRenew handles token renewal requests (POST /renew)
// Requires refresh token validation (handled by middleware)
// Forwards both service JWT (for auth) and original refresh token (for DB validation)
func (h *ProxyHandler) ProxyRenew(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyRenewWithServiceJWT(w, r, h.logger, h.userDBClient.ForwardWithServiceJWT)
}

// ProxyUserDatabase handles all user database service routes
// Requires normal JWT validation (handled by middleware)
func (h *ProxyHandler) ProxyUserDatabase(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.logger, h.userDBClient.ForwardWithServiceJWT)
}

// ProxyRecommendation handles all recommendation service routes
// Requires normal JWT validation (handled by middleware)
func (h *ProxyHandler) ProxyRecommendation(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.logger, h.recommendClient.ForwardWithServiceJWT)
}

// ProxyEventIngestion handles all event ingestion service routes
// Requires normal JWT validation (handled by middleware)
func (h *ProxyHandler) ProxyEventIngestion(w http.ResponseWriter, r *http.Request) {
	libshandlers.ProxyWithServiceJWT(w, r, h.logger, h.eventIngestionClient.ForwardWithServiceJWT)
}
