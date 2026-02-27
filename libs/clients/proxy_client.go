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

// ProxyClient is a generic HTTP client for proxying requests to backend services
type ProxyClient struct {
	BaseClient
	baseURL string
}

// NewProxyClient creates a new proxy client with the given base URL
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

// ForwardRequest forwards an HTTP request to the backend service
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

// ForwardWithServiceJWT forwards an HTTP request with a service JWT in the Authorization header
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

	// Create a copy of headers to avoid modifying the original
	headersCopy := make(http.Header)
	for key, values := range headers {
		// Skip the original Authorization header
		if key == "Authorization" {
			continue
		}
		for _, value := range values {
			headersCopy.Add(key, value)
		}
	}

	// Add service JWT as Authorization header
	headersCopy.Set("Authorization", "Bearer "+serviceJWT)

	return c.BaseClient.DoProxy(ctx, method, fullURL, body, headersCopy, requestID)
}

// buildURL constructs the full URL from base URL, path, and query string
func (c *ProxyClient) buildURL(path, query string) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base URL: %w", err)
	}

	u.Path = path
	u.RawQuery = query

	return u.String(), nil
}

// ForwardRequestWithBuffer is a convenience method that accepts a byte buffer
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

// ForwardWithServiceJWTBuffer is a convenience method that accepts a byte buffer
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
