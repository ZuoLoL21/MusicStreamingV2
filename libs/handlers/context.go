package handlers

import (
	"context"

	libsdi "libs/di"
)

// GetRequestIDFromContext extracts request ID from context (libsdi.RequestIDKey)
//
// Returns empty string if not found
func GetRequestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(libsdi.RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// GetServiceJWTFromContext extracts service JWT from context (libsdi.ServiceJWTKey)
//
// Returns empty string if not found
func GetServiceJWTFromContext(ctx context.Context) string {
	if jwt, ok := ctx.Value(libsdi.ServiceJWTKey).(string); ok {
		return jwt
	}
	return ""
}

// GetUserUUIDFromContext extracts user UUID from context (using libsdi.UserUUIDKey)
//
// Returns empty string if not found
func GetUserUUIDFromContext(ctx context.Context) string {
	if uuid, ok := ctx.Value(libsdi.UserUUIDKey).(string); ok {
		return uuid
	}
	return ""
}
