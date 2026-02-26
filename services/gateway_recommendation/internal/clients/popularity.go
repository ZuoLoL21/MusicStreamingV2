package clients

import (
	"context"
	"fmt"
	"net/http"
	"time"

	libsclients "libs/clients"

	"go.uber.org/zap"
)

type PopularityClient struct {
	libsclients.BaseClient
	baseURL string
}

type ThemePopularity struct {
	Theme              string  `json:"theme"`
	DecayPlays         float64 `json:"decay_plays"`
	DecayListenSeconds float64 `json:"decay_listen_seconds"`
}

func NewPopularityClient(baseURL string, logger *zap.Logger) *PopularityClient {
	return &PopularityClient{
		BaseClient: libsclients.BaseClient{
			HttpClient: &http.Client{Timeout: 15 * time.Second},
			Logger:     logger,
		},
		baseURL: baseURL,
	}
}

func (c *PopularityClient) GetThemePopularity(ctx context.Context, requestID string, limit int) ([]ThemePopularity, error) {
	c.Logger.Info("Fetching theme popularity",
		zap.String("request_id", requestID),
		zap.Int("limit", limit))

	var themes []ThemePopularity
	url := fmt.Sprintf("%s/popular/themes/all-time?limit=%d", c.baseURL, limit)

	if err := c.DoJSON(ctx, "GET", url, nil, &themes, requestID); err != nil {
		c.Logger.Error("Theme popularity fetch failed",
			zap.String("request_id", requestID),
			zap.Error(err))
		return nil, fmt.Errorf("fetch theme popularity: %w", err)
	}

	c.Logger.Info("Theme popularity fetched",
		zap.String("request_id", requestID),
		zap.Int("count", len(themes)))

	return themes, nil
}

func (c *PopularityClient) ProxyRequest(ctx context.Context, method string, path string, queryParams string, requestID string) ([]byte, int, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)
	if queryParams != "" {
		url = fmt.Sprintf("%s?%s", url, queryParams)
	}

	c.Logger.Info("Proxying request",
		zap.String("request_id", requestID),
		zap.String("method", method),
		zap.String("path", path))

	body, statusCode, err := c.DoRaw(ctx, method, url, requestID)
	if err != nil {
		c.Logger.Error("Proxy request failed",
			zap.String("request_id", requestID),
			zap.Error(err))
		return body, statusCode, fmt.Errorf("proxy request: %w", err)
	}

	c.Logger.Info("Proxy request completed",
		zap.String("request_id", requestID),
		zap.Int("status_code", statusCode))

	return body, statusCode, nil
}
