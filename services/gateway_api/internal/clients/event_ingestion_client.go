package clients

import (
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
