package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	libsclients "libs/clients"

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

func NewPopularityClient(baseURL string, logger *zap.Logger) *PopularityClient {
	return &PopularityClient{
		ProxyClient: libsclients.NewProxyClient(baseURL, logger),
	}
}

func (c *PopularityClient) GetThemePopularity(ctx context.Context, requestID string, serviceJWT string, limit int) ([]ThemePopularity, error) {
	c.Logger.Info("Fetching theme popularity",
		zap.String("request_id", requestID),
		zap.Int("limit", limit))

	var themes []ThemePopularity
	url := fmt.Sprintf("/popular/themes/all-time?limit=%d", limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Request-ID", requestID)
	req.Header.Set("Authorization", "Bearer "+serviceJWT)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.Logger.Error("Theme popularity fetch failed",
			zap.String("request_id", requestID),
			zap.Error(err))
		return nil, fmt.Errorf("fetch theme popularity: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.Logger.Error("Theme popularity fetch failed with non-200 status",
			zap.String("request_id", requestID),
			zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&themes); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.Logger.Info("Theme popularity fetched",
		zap.String("request_id", requestID),
		zap.Int("count", len(themes)))

	return themes, nil
}
