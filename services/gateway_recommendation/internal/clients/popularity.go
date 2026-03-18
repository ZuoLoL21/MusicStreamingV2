package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	libsclients "libs/clients"
	libsmiddleware "libs/middleware"

	"go.uber.org/zap"
)

type PopularityClient struct {
	*libsclients.ProxyClient
}

type ThemePopularity struct {
	Theme              string  `json:"theme"`
	DecayPlays         float64 `json:"decay_plays"`
	DecayListenSeconds float64 `json:"decay_listen_seconds"`
}

func NewPopularityClient(baseURL string) *PopularityClient {
	return &PopularityClient{
		ProxyClient: libsclients.NewProxyClient(baseURL),
	}
}

func (c *PopularityClient) GetThemePopularity(ctx context.Context, requestID string, serviceJWT string, limit int) ([]ThemePopularity, error) {
	logger := libsmiddleware.GetLogger(ctx)

	var themes []ThemePopularity
	path := "/popular/themes/all-time"
	query := fmt.Sprintf("limit=%d", limit)

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	respBody, statusCode, _, err := c.ForwardWithServiceJWT(
		ctx,
		"GET",
		path,
		query,
		nil,
		headers,
		serviceJWT,
		requestID,
	)
	if err != nil {
		logger.Error("Theme popularity fetch failed", zap.Error(err))
		return nil, fmt.Errorf("fetch theme popularity: %w", err)
	}

	if statusCode != http.StatusOK {
		logger.Error("Theme popularity fetch failed with non-200 status",
			zap.Int("status_code", statusCode),
			zap.String("response", string(respBody)))
		return nil, fmt.Errorf("unexpected status code: %d", statusCode)
	}

	if err := json.Unmarshal(respBody, &themes); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	logger.Info("Theme popularity fetched", zap.Int("count", len(themes)))

	return themes, nil
}
