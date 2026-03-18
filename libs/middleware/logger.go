package middleware

import (
	"context"
	"libs/helpers"
	"net/http"

	"go.uber.org/zap"
)

// loggerKeyType is a private type used as a key for storing the logger in context.
// Using an unexported type prevents collisions with other context keys.
type loggerKeyType struct{}

// loggerKey is the context key used to store and retrieve the zap.Logger.
// This unexported key prevents accidental collisions with other context values.
var loggerKey = loggerKeyType{}

// Logger returns a middleware that adds a zap.Logger to the request context.
// The logger is enriched with request metadata (method, path) and context values
// (requestID, userUUID, deviceID) if present. The logger can be retrieved using GetLogger.
//
// The logger is stored in context using a private key to avoid collisions.
func Logger(baseLogger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := baseLogger

			if requestID := helpers.GetRequestIDFromContext(r.Context()); requestID != "" {
				logger = logger.With(zap.String("request_id", requestID))
			}
			if userUUID := helpers.GetUserUUIDFromContext(r.Context()); userUUID != "" {
				logger = logger.With(zap.String("user_uuid", userUUID))
			}
			if deviceID := helpers.GetDeviceIDFromContext(r.Context()); deviceID != "" {
				logger = logger.With(zap.String("device_uuid", deviceID))
			}

			logger = logger.With(
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			ctx := context.WithValue(r.Context(), loggerKey, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetLogger retrieves the zap.Logger from the request context.
// Returns a no-op logger if no logger was added to the context.
// Use this to get the logger that was set up by the Logger middleware.
func GetLogger(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	return zap.NewNop()
}
