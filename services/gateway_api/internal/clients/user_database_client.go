package clients

import (
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
