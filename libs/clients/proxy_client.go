package clients

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// ProxyClient is a generic HTTP client for proxying requests to backend services.
// It embeds BaseClient and adds functionality for forwarding requests with optional
// service JWT authentication. The client maintains a base URL for the backend service.
type ProxyClient struct {
	BaseClient
	baseURL string
}

// NewProxyClient creates a new proxy client with the given base URL.
//
//   - The baseURL should be the full URL of the backend service (e.g., "http://localhost:8080").
//   - The logger is used for logging HTTP requests and responses.
//
// Returns a configured ProxyClient with a 30-second timeout.
func NewProxyClient(baseURL string, logger *zap.Logger) *ProxyClient {
	return &ProxyClient{
		BaseClient: BaseClient{
			HttpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
			Logger: logger,
		},
		baseURL: baseURL,
	}
}

// ForwardRequest forwards an HTTP request to the backend service.
//
// It takes the HTTP method, path (appended to baseURL), optional query string,
// body reader, headers, and request ID. Returns the response body, status code,
// and response headers.
//
// The Authorization header is excluded from forwarded headers.
func (c *ProxyClient) ForwardRequest(
	ctx context.Context,
	method, path, query string,
	body io.Reader,
	headers http.Header,
	requestID string,
) ([]byte, int, http.Header, error) {
	fullURL, err := c.buildURL(path, query)
	if err != nil {
		return nil, http.StatusInternalServerError, nil, fmt.Errorf("build URL: %w", err)
	}

	return c.BaseClient.DoProxy(ctx, method, fullURL, body, headers, requestID)
}

// ForwardWithServiceJWT forwards an HTTP request with a service JWT in the Authorization header.
//
// This method adds a Bearer token with the provided serviceJWT to the request headers
// before forwarding. It excludes any existing Authorization header from the original request.
//
// Useful for authenticated service-to-service communication.
func (c *ProxyClient) ForwardWithServiceJWT(
	ctx context.Context,
	method, path, query string,
	body io.Reader,
	headers http.Header,
	serviceJWT, requestID string,
) ([]byte, int, http.Header, error) {
	fullURL, err := c.buildURL(path, query)
	if err != nil {
		return nil, http.StatusInternalServerError, nil, fmt.Errorf("build URL: %w", err)
	}

	// Manage headers
	headersCopy := make(http.Header)
	for key, values := range headers {
		if key == "Authorization" {
			continue
		}
		for _, value := range values {
			headersCopy.Add(key, value)
		}
	}
	headersCopy.Set("Authorization", "Bearer "+serviceJWT)

	return c.BaseClient.DoProxy(ctx, method, fullURL, body, headersCopy, requestID)
}

// buildURL constructs the full URL from base URL, path, and query string.
// It parses the base URL, appends the given path, and sets the query string parameters.
//
// Returns the complete URL as a string, or an error if URL parsing fails.
func (c *ProxyClient) buildURL(path, query string) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base URL: %w", err)
	}

	u.Path = path
	u.RawQuery = query

	return u.String(), nil
}

// ForwardRequestWithBuffer is a convenience method that accepts a byte buffer instead of an io.Reader.
// It converts the bodyBytes to an io.Reader and calls ForwardRequest.
//
// If bodyBytes is empty, no body is sent with the request.
func (c *ProxyClient) ForwardRequestWithBuffer(
	ctx context.Context,
	method, path, query string,
	bodyBytes []byte,
	headers http.Header,
	requestID string,
) ([]byte, int, http.Header, error) {
	var body io.Reader
	if len(bodyBytes) > 0 {
		body = bytes.NewReader(bodyBytes)
	}
	return c.ForwardRequest(ctx, method, path, query, body, headers, requestID)
}

// ForwardWithServiceJWTBuffer is a convenience method that accepts a byte buffer instead of an io.Reader.
// It converts the bodyBytes to an io.Reader and calls ForwardWithServiceJWT.
//
// If bodyBytes is empty, no body is sent with the request.
func (c *ProxyClient) ForwardWithServiceJWTBuffer(
	ctx context.Context,
	method, path, query string,
	bodyBytes []byte,
	headers http.Header,
	serviceJWT, requestID string,
) ([]byte, int, http.Header, error) {
	var body io.Reader
	if len(bodyBytes) > 0 {
		body = bytes.NewReader(bodyBytes)
	}
	return c.ForwardWithServiceJWT(ctx, method, path, query, body, headers, serviceJWT, requestID)
}
