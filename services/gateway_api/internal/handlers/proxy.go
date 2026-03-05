package handlers

import (
	"bytes"
	"gateway_api/internal/clients"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// ProxyHandler handles proxying requests to backend services
type ProxyHandler struct {
	userDBClient         *clients.UserDatabaseClient
	recommendClient      *clients.RecommendationClient
	eventIngestionClient *clients.EventIngestionClient
	logger               *zap.Logger
	requestIDKey         any
	serviceJWTKey        any
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(
	userDBClient *clients.UserDatabaseClient,
	recommendClient *clients.RecommendationClient,
	eventIngestionClient *clients.EventIngestionClient,
	logger *zap.Logger,
	requestIDKey any,
	serviceJWTKey any,
) *ProxyHandler {
	return &ProxyHandler{
		userDBClient:         userDBClient,
		recommendClient:      recommendClient,
		eventIngestionClient: eventIngestionClient,
		logger:               logger,
		requestIDKey:         requestIDKey,
		serviceJWTKey:        serviceJWTKey,
	}
}

// ProxyLogin handles login requests (POST and PUT /login)
// These are public routes that don't require authentication
func (h *ProxyHandler) ProxyLogin(w http.ResponseWriter, r *http.Request) {
	requestID := h.extractRequestID(r)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read login request body",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = r.Body.Close()

	headers := h.copyHeaders(r.Header, false)
	respBody, statusCode, respHeaders, err := h.userDBClient.ForwardRequest(
		r.Context(),
		r.Method,
		r.URL.Path,
		r.URL.RawQuery,
		bytes.NewReader(bodyBytes),
		headers,
		requestID,
	)

	if err != nil {
		h.logger.Error("failed to forward login request to backend",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method))
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}

	h.writeResponse(w, respBody, statusCode, respHeaders)
}

// ProxyRenew handles token renewal requests (POST /renew)
// Requires refresh token validation (handled by middleware)
func (h *ProxyHandler) ProxyRenew(w http.ResponseWriter, r *http.Request) {
	requestID := h.extractRequestID(r)
	serviceJWT := h.extractServiceJWT(r)

	if serviceJWT == "" {
		h.logger.Error("service JWT not found in context for renew request",
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read renew request body",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = r.Body.Close()

	headers := h.copyHeaders(r.Header, true)

	respBody, statusCode, respHeaders, err := h.userDBClient.ProxyRequest(
		r.Context(),
		r.Method,
		r.URL.Path,
		r.URL.RawQuery,
		bytes.NewReader(bodyBytes),
		headers,
		serviceJWT,
		requestID,
	)

	if err != nil {
		h.logger.Error("failed to forward renew request to backend",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method))
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}

	h.writeResponse(w, respBody, statusCode, respHeaders)
}

// ProxyUserDatabase handles all user database service routes
// Requires normal JWT validation (handled by middleware)
func (h *ProxyHandler) ProxyUserDatabase(w http.ResponseWriter, r *http.Request) {
	requestID := h.extractRequestID(r)
	serviceJWT := h.extractServiceJWT(r)

	if serviceJWT == "" {
		h.logger.Error("service JWT not found in context for user database request",
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read user database request body",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = r.Body.Close()

	headers := h.copyHeaders(r.Header, true)

	respBody, statusCode, respHeaders, err := h.userDBClient.ProxyRequest(
		r.Context(),
		r.Method,
		r.URL.Path,
		r.URL.RawQuery,
		bytes.NewReader(bodyBytes),
		headers,
		serviceJWT,
		requestID,
	)

	if err != nil {
		h.logger.Error("failed to forward user database request to backend",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method))
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}

	h.writeResponse(w, respBody, statusCode, respHeaders)
}

// ProxyRecommendation handles all recommendation service routes
// Requires normal JWT validation (handled by middleware)
func (h *ProxyHandler) ProxyRecommendation(w http.ResponseWriter, r *http.Request) {
	requestID := h.extractRequestID(r)
	serviceJWT := h.extractServiceJWT(r)

	if serviceJWT == "" {
		h.logger.Error("service JWT not found in context for recommendation request",
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read recommendation request body",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = r.Body.Close()

	headers := h.copyHeaders(r.Header, true)

	respBody, statusCode, respHeaders, err := h.recommendClient.ProxyRequest(
		r.Context(),
		r.Method,
		r.URL.Path,
		r.URL.RawQuery,
		bytes.NewReader(bodyBytes),
		headers,
		serviceJWT,
		requestID,
	)

	if err != nil {
		h.logger.Error("failed to forward recommendation request to backend",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method))
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}

	h.writeResponse(w, respBody, statusCode, respHeaders)
}

// ProxyEventIngestion handles all event ingestion service routes
// Requires normal JWT validation (handled by middleware)
func (h *ProxyHandler) ProxyEventIngestion(w http.ResponseWriter, r *http.Request) {
	requestID := h.extractRequestID(r)
	serviceJWT := h.extractServiceJWT(r)

	if serviceJWT == "" {
		h.logger.Error("service JWT not found in context for event ingestion request",
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read event ingestion request body",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = r.Body.Close()

	headers := h.copyHeaders(r.Header, true)

	respBody, statusCode, respHeaders, err := h.eventIngestionClient.ProxyRequest(
		r.Context(),
		r.Method,
		r.URL.Path,
		r.URL.RawQuery,
		bytes.NewReader(bodyBytes),
		headers,
		serviceJWT,
		requestID,
	)

	if err != nil {
		h.logger.Error("failed to forward event ingestion request to backend",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method))
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}

	h.writeResponse(w, respBody, statusCode, respHeaders)
}

// extractRequestID extracts the request ID from the context
func (h *ProxyHandler) extractRequestID(r *http.Request) string {
	requestID, ok := r.Context().Value(h.requestIDKey).(string)
	if !ok {
		return ""
	}
	return requestID
}

// extractServiceJWT extracts the service JWT from the context
func (h *ProxyHandler) extractServiceJWT(r *http.Request) string {
	serviceJWT, ok := r.Context().Value(h.serviceJWTKey).(string)
	if !ok {
		return ""
	}
	return serviceJWT
}

// copyHeaders copies HTTP headers, optionally excluding Authorization
func (h *ProxyHandler) copyHeaders(headers http.Header, excludeAuth bool) http.Header {
	copied := make(http.Header)
	for key, values := range headers {
		if excludeAuth && key == "Authorization" {
			continue
		}
		for _, value := range values {
			copied.Add(key, value)
		}
	}
	return copied
}

// writeResponse writes the response body and status code
func (h *ProxyHandler) writeResponse(w http.ResponseWriter, body []byte, statusCode int, headers http.Header) {
	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}
