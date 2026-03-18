package handlers

import (
	"bytes"
	"context"
	"io"
	"libs/consts"
	"libs/helpers"
	"net/http"

	"go.uber.org/zap"
)

// CopyHeaders copies HTTP headers from the source to a new header map.
//
// If excludeAuth is true, the "Authorization" header is excluded from the copy.
//
// This is useful for forwarding headers to backend services while optionally
// removing authentication headers.
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

// WriteProxyResponse writes the proxy response to the client.
//
// It copies all headers from the backend response to the client response,
// sets the status code, and writes the response body.
//
// This is the final step in proxying a response back to the original client.
func WriteProxyResponse(w http.ResponseWriter, body []byte, statusCode int, headers http.Header) {
	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

// ProxyWithServiceJWT handles authenticated proxy requests using service JWT.
//
// This handler extracts the service JWT from the request context, validates it exists,
// reads the request body, forwards the request with the JWT to the backend service,
// and writes the response back to the client.
//
// It returns 401 Unauthorized if the service JWT is missing from context.
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
		http.Error(w, consts.ErrUnauthorized, http.StatusUnauthorized)
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

// ProxyRenewWithServiceJWT handles token renewal by forwarding both the service JWT
// (in Authorization header) and the original refresh token (in X-Refresh-Token header).
// This allows the backend to authenticate via service JWT while validating the refresh token hash.
//
// It returns 401 Unauthorized if the service JWT is missing from context.
func ProxyRenewWithServiceJWT(
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
		http.Error(w, consts.ErrUnauthorized, http.StatusUnauthorized)
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
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		headers.Set("X-Refresh-Token", authHeader)
	}

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

// ProxyPublic handles public proxy requests with no JWT processing.
//
// This handler forwards all headers (including Authorization) to the backend service
// as-is.
//
// It reads the request body, forwards it to the backend, and writes
// the response back to the client. Use this for routes that should not require
// authentication or should pass through the original credentials.
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
