package clients

import (
	"context"
	"io"
	"net/http"

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

// ProxyRequest forwards a request to the recommendation service with a service JWT
func (c *RecommendationClient) ProxyRequest(
	ctx context.Context,
	method, path, query string,
	body io.Reader,
	headers http.Header,
	serviceJWT, requestID string,
) ([]byte, int, error) {
	return c.ForwardWithServiceJWT(ctx, method, path, query, body, headers, serviceJWT, requestID)
}
