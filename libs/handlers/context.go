package handlers

import (
	"context"
	"libs/consts"
)

// GetRequestIDFromContext extracts request ID from context (consts.RequestIDKey)
//
// Returns empty string if not found
func GetRequestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(consts.RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// GetServiceJWTFromContext extracts service JWT from context (consts.ServiceJWTKey)
//
// Returns empty string if not found
func GetServiceJWTFromContext(ctx context.Context) string {
	if jwt, ok := ctx.Value(consts.ServiceJWTKey).(string); ok {
		return jwt
	}
	return ""
}

// GetUserUUIDFromContext extracts user UUID from context (using consts.UserUUIDKey)
//
// Returns empty string if not found
func GetUserUUIDFromContext(ctx context.Context) string {
	if uuid, ok := ctx.Value(consts.UserUUIDKey).(string); ok {
		return uuid
	}
	return ""
}
