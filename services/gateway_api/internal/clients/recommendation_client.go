package clients

import (
	libsclients "libs/clients"

	"go.uber.org/zap"
)

// RecommendationClient is a client for the recommendation service
type RecommendationClient struct {
	*libsclients.ProxyClient
}

// NewRecommendationClient creates a new recommendation client
func NewRecommendationClient(baseURL string, logger *zap.Logger) *RecommendationClient {
	return &RecommendationClient{
		ProxyClient: libsclients.NewProxyClient(baseURL, logger),
	}
}
