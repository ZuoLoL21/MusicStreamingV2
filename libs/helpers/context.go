package helpers

import (
	"context"
	"libs/consts"
)

// GetRequestIDFromContext extracts the request ID from the given context.
// The request ID is stored in the context using the key defined by consts.RequestIDKey.
// Returns an empty string if the request ID is not found in the context.
func GetRequestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(consts.RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// GetServiceJWTFromContext extracts the service JWT from the given context.
// The service JWT is stored in the context using the key defined by consts.ServiceJWTKey.
// Returns an empty string if the service JWT is not found in the context.
func GetServiceJWTFromContext(ctx context.Context) string {
	if jwt, ok := ctx.Value(consts.ServiceJWTKey).(string); ok {
		return jwt
	}
	return ""
}

// GetUserUUIDFromContext extracts the user UUID from the given context.
// The user UUID is stored in the context using the key defined by consts.UserUUIDKey.
// Returns an empty string if the user UUID is not found in the context.
func GetUserUUIDFromContext(ctx context.Context) string {
	if uuid, ok := ctx.Value(consts.UserUUIDKey).(string); ok {
		return uuid
	}
	return ""
}

// GetDeviceIDFromContext extracts the device ID from the given context.
// The device ID is stored in the context using the key defined by consts.DeviceIDKey.
// Returns an empty string if the device ID is not found in the context.
func GetDeviceIDFromContext(ctx context.Context) string {
	if deviceID, ok := ctx.Value(consts.DeviceIDKey).(string); ok {
		return deviceID
	}
	return ""
}
