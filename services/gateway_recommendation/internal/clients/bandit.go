package clients

import (
	"context"
	"fmt"
	"net/http"
	"time"

	libsclients "libs/clients"

	"go.uber.org/zap"
)

type BanditClient struct {
	libsclients.BaseClient
	baseURL string
}

func NewBanditClient(baseURL string, logger *zap.Logger) *BanditClient {
	return &BanditClient{
		BaseClient: libsclients.BaseClient{
			HttpClient: &http.Client{Timeout: 30 * time.Second},
			Logger:     logger,
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
	c.Logger.Info("Calling bandit service",
		zap.String("request_id", requestID),
		zap.String("user_uuid", userUUID))

	req := PredictRequest{UserUUID: userUUID}
	var resp PredictResponse

	url := fmt.Sprintf("%s/api/v1/predict", c.baseURL)
	if err := c.DoJSON(ctx, "POST", url, req, &resp, requestID); err != nil {
		c.Logger.Error("Bandit prediction failed",
			zap.String("request_id", requestID),
			zap.Error(err))
		return nil, fmt.Errorf("bandit prediction: %w", err)
	}

	c.Logger.Info("Bandit prediction successful",
		zap.String("request_id", requestID),
		zap.String("theme", resp.Theme))

	return &resp, nil
}
