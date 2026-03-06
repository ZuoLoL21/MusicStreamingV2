package handlers

import (
	"bytes"
	"context"
	"io"
	"libs/helpers"
	"net/http"

	"go.uber.org/zap"
)

// CopyHeaders copies HTTP headers, optionally excluding Authorization
func CopyHeaders(headers http.Header, excludeAuth bool) http.Header {
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

// WriteProxyResponse writes proxy response with headers and body
func WriteProxyResponse(w http.ResponseWriter, body []byte, statusCode int, headers http.Header) {
	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

// ProxyWithServiceJWT handles standard authenticated proxy pattern
// Extracts context values → validates JWT → reads body → forwards → writes response
func ProxyWithServiceJWT(
	w http.ResponseWriter,
	r *http.Request,
	logger *zap.Logger,
	forwardFunc func(ctx context.Context, method, path, query string, body io.Reader, headers http.Header, serviceJWT, requestID string) ([]byte, int, http.Header, error),
) {
	requestID := helpers.GetRequestIDFromContext(r.Context())
	serviceJWT := helpers.GetServiceJWTFromContext(r.Context())

	if serviceJWT == "" {
		logger.Error("service JWT not found in context",
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("failed to read request body",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = r.Body.Close()

	headers := CopyHeaders(r.Header, true)

	respBody, statusCode, respHeaders, err := forwardFunc(
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
		logger.Error("failed to forward request",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method))
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}

	WriteProxyResponse(w, respBody, statusCode, respHeaders)
}

// ProxyPublic handles truly public routes with no JWT processing
// Forwards all headers as-is → reads body → forwards → writes response
func ProxyPublic(
	w http.ResponseWriter,
	r *http.Request,
	logger *zap.Logger,
	forwardFunc func(ctx context.Context, method, path, query string, body io.Reader, headers http.Header, requestID string) ([]byte, int, http.Header, error),
) {
	requestID := helpers.GetRequestIDFromContext(r.Context())

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("failed to read request body",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = r.Body.Close()

	headers := CopyHeaders(r.Header, false) // Keep all headers including auth

	respBody, statusCode, respHeaders, err := forwardFunc(
		r.Context(),
		r.Method,
		r.URL.Path,
		r.URL.RawQuery,
		bytes.NewReader(bodyBytes),
		headers,
		requestID,
	)

	if err != nil {
		logger.Error("failed to forward request",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}

	WriteProxyResponse(w, respBody, statusCode, respHeaders)
}
