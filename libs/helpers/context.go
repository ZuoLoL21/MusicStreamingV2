package helpers

import (
	"context"
	"libs/consts"
)

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(consts.RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// GetServiceJWTFromContext extracts service JWT from context
func GetServiceJWTFromContext(ctx context.Context) string {
	if jwt, ok := ctx.Value(consts.ServiceJWTKey).(string); ok {
		return jwt
	}
	return ""
}

// GetUserUUIDFromContext extracts user UUID from context
func GetUserUUIDFromContext(ctx context.Context) string {
	if uuid, ok := ctx.Value(consts.UserUUIDKey).(string); ok {
		return uuid
	}
	return ""
}
