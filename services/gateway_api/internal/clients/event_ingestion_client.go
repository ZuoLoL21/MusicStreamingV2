package clients

import (
	"context"
	"io"
	"net/http"

	libsclients "libs/clients"

	"go.uber.org/zap"
)

// EventIngestionClient is a client for the event ingestion service
type EventIngestionClient struct {
	*libsclients.ProxyClient
}

// NewEventIngestionClient creates a new event ingestion client
func NewEventIngestionClient(baseURL string, logger *zap.Logger) *EventIngestionClient {
	return &EventIngestionClient{
		ProxyClient: libsclients.NewProxyClient(baseURL, logger),
	}
}

// ProxyRequest forwards a request to the event ingestion service with a service JWT
func (c *EventIngestionClient) ProxyRequest(
	ctx context.Context,
	method, path, query string,
	body io.Reader,
	headers http.Header,
	serviceJWT, requestID string,
) ([]byte, int, http.Header, error) {
	return c.ForwardWithServiceJWT(ctx, method, path, query, body, headers, serviceJWT, requestID)
}
