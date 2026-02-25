package clients

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type BanditClient struct {
	baseClient
	baseURL string
}

func NewBanditClient(baseURL string, logger *zap.Logger) *BanditClient {
	return &BanditClient{
		baseClient: baseClient{
			httpClient: &http.Client{Timeout: 30 * time.Second},
			logger:     logger,
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
	c.logger.Info("Calling bandit service",
		zap.String("request_id", requestID),
		zap.String("user_uuid", userUUID))

	req := PredictRequest{UserUUID: userUUID}
	var resp PredictResponse

	url := fmt.Sprintf("%s/api/v1/predict", c.baseURL)
	if err := c.doJSON(ctx, "POST", url, req, &resp, requestID); err != nil {
		c.logger.Error("Bandit prediction failed",
			zap.String("request_id", requestID),
			zap.Error(err))
		return nil, fmt.Errorf("bandit prediction: %w", err)
	}

	c.logger.Info("Bandit prediction successful",
		zap.String("request_id", requestID),
		zap.String("theme", resp.Theme))

	return &resp, nil
}
