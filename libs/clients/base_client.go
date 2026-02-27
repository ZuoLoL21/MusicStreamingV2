package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// BaseClient provides common HTTP functionality for service clients
type BaseClient struct {
	HttpClient *http.Client
	Logger     *zap.Logger
}

// DoJSON performs an HTTP request with JSON marshaling/unmarshaling
func (b *BaseClient) DoJSON(ctx context.Context, method, url string, reqBody, respBody interface{}, requestID string) error {
	var bodyReader io.Reader
	if reqBody != nil {
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Request-ID", requestID)

	resp, err := b.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

// DoRaw performs an HTTP request and returns raw response
func (b *BaseClient) DoRaw(ctx context.Context, method, url string, requestID string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Request-ID", requestID)

	resp, err := b.HttpClient.Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("read response: %w", err)
	}

	return body, resp.StatusCode, nil
}

// DoProxy forwards a raw HTTP request with body and headers, returning raw response with headers
func (b *BaseClient) DoProxy(ctx context.Context, method, url string, body io.Reader, headers http.Header, requestID string) ([]byte, int, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, http.StatusInternalServerError, nil, fmt.Errorf("create request: %w", err)
	}

	// Copy headers from original request
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Ensure request ID is set
	req.Header.Set("X-Request-ID", requestID)

	resp, err := b.HttpClient.Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, nil, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, nil, fmt.Errorf("read response: %w", err)
	}

	return respBody, resp.StatusCode, resp.Header, nil
}
