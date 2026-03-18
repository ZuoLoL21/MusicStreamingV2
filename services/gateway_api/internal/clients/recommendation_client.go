package clients

import (
	libsclients "libs/clients"
)

// RecommendationClient is a client for the recommendation service
type RecommendationClient struct {
	*libsclients.ProxyClient
}

// NewRecommendationClient creates a new recommendation client
func NewRecommendationClient(baseURL string) *RecommendationClient {
	return &RecommendationClient{
		ProxyClient: libsclients.NewProxyClient(baseURL),
	}
}
