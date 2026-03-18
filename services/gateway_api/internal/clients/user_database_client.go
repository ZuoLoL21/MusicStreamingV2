package clients

import (
	libsclients "libs/clients"
)

// UserDatabaseClient is a client for the user database service
type UserDatabaseClient struct {
	*libsclients.ProxyClient
}

// NewUserDatabaseClient creates a new user database client
func NewUserDatabaseClient(baseURL string) *UserDatabaseClient {
	return &UserDatabaseClient{
		ProxyClient: libsclients.NewProxyClient(baseURL),
	}
}
