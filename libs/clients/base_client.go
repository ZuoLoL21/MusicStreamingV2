package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
	"libs/consts"
)

// BaseClient provides common HTTP functionality for service clients.
//
// It wraps an http.Client and a zap.Logger for making HTTP requests
// with built-in error handling and JSON marshaling support.
type BaseClient struct {
	HttpClient *http.Client
	Logger     *zap.Logger
}

// DoJSON performs an HTTP request with JSON marshaling/unmarshaling.
//
// It takes the HTTP method, URL, optional request body (will be JSON-marshaled),
// a response body (will be JSON-unmarshaled into), and a request ID for tracking.
//
// Returns an error if the request fails, status code is not 200, or JSON marshaling fails.
func (b *BaseClient) DoJSON(ctx context.Context, method, url string, reqBody, respBody interface{}, requestID string) error {
	var bodyReader io.Reader
	if reqBody != nil {
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("%s: %w", consts.ErrMarshalRequest, err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("%s: %w", consts.ErrCreateRequest, err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Request-ID", requestID)

	resp, err := b.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", consts.ErrRequestFailed, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: %w", consts.ErrReadResponse, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("%s: %w", consts.ErrUnmarshalResponse, err)
		}
	}

	return nil
}

// DoRaw performs an HTTP request and returns raw response bytes.
//
// Unlike DoJSON, this method does not perform any JSON marshaling/unmarshaling.
// It returns the response body as []byte along with the HTTP status code.
// Useful for endpoints that return non-JSON responses or when raw body handling is needed.
func (b *BaseClient) DoRaw(ctx context.Context, method, url string, requestID string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("%s: %w", consts.ErrCreateRequest, err)
	}

	req.Header.Set("X-Request-ID", requestID)

	resp, err := b.HttpClient.Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("%s: %w", consts.ErrRequestFailed, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("%s: %w", consts.ErrReadResponse, err)
	}

	return body, resp.StatusCode, nil
}

// DoProxy forwards an HTTP request with full control over body and headers.
//
// This is the most flexible method - it accepts a body reader and allows passing
// through custom headers. It returns the raw response body, status code, and response headers.
// Used primarily for proxying requests where headers need to be preserved or modified.
func (b *BaseClient) DoProxy(ctx context.Context, method, url string, body io.Reader, headers http.Header, requestID string) ([]byte, int, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, http.StatusInternalServerError, nil, fmt.Errorf("%s: %w", consts.ErrCreateRequest, err)
	}

	// Headers
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Header.Set("X-Request-ID", requestID)

	// Call
	resp, err := b.HttpClient.Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, nil, fmt.Errorf("%s: %w", consts.ErrRequestFailed, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, nil, fmt.Errorf("%s: %w", consts.ErrReadResponse, err)
	}

	return respBody, resp.StatusCode, resp.Header, nil
}
