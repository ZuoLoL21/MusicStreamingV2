package clients

import (
	"context"
	"io"
	"net/http"

	libsclients "libs/clients"

	"go.uber.org/zap"
)

// UserDatabaseClient is a client for the user database service
type UserDatabaseClient struct {
	*libsclients.ProxyClient
}

// NewUserDatabaseClient creates a new user database client
func NewUserDatabaseClient(baseURL string, logger *zap.Logger) *UserDatabaseClient {
	return &UserDatabaseClient{
		ProxyClient: libsclients.NewProxyClient(baseURL, logger),
	}
}

// ProxyRequest forwards a request to the user database service with a service JWT
func (c *UserDatabaseClient) ProxyRequest(
	ctx context.Context,
	method, path, query string,
	body io.Reader,
	headers http.Header,
	serviceJWT, requestID string,
) ([]byte, int, error) {
	return c.ForwardWithServiceJWT(ctx, method, path, query, body, headers, serviceJWT, requestID)
}
