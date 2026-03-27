package clients

import (
	"context"
	"fmt"
	"libs/metrics"
	"net/http"
	"time"

	libsclients "libs/clients"
	libsmiddleware "libs/middleware"

	"go.uber.org/zap"
)

type BanditClient struct {
	libsclients.BaseClient
	baseURL string
}

func NewBanditClient(baseURL string) *BanditClient {
	return &BanditClient{
		BaseClient: libsclients.BaseClient{
			HttpClient: &http.Client{Timeout: 30 * time.Second},
		},
		baseURL: baseURL,
	}
}

type PredictRequest struct {
	UserUUID string `json:"user_uuid"`
}

type PredictResponse struct {
	Theme    string    `json:"theme"`
	Features []float64 `json:"features"`
}

func (c *BanditClient) Predict(ctx context.Context, userUUID string, requestID string) (*PredictResponse, error) {
	logger := libsmiddleware.GetLogger(ctx)

	req := PredictRequest{UserUUID: userUUID}
	var resp PredictResponse

	url := fmt.Sprintf("%s/api/v1/predict", c.baseURL)

	start := time.Now()
	err := c.DoJSON(ctx, "POST", url, req, &resp, requestID)
	metrics.TrackDownstreamCall("bandit", "/predict", time.Since(start), err)

	if err != nil {
		logger.Error("Bandit prediction failed", zap.Error(err))
		return nil, fmt.Errorf("bandit prediction: %w", err)
	}

	logger.Info("Bandit prediction successful", zap.String("theme", resp.Theme))

	return &resp, nil
}
