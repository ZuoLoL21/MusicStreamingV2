package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

// LoggerConfig holds configuration for the logger middleware
type LoggerConfig struct {
	RequestIDKey any
	UserUUIDKey  any
}

// Logger creates a request-scoped logger with context fields pre-bound
// This middleware should be placed AFTER RequestID middleware and BEFORE/AFTER Auth middleware
func Logger(baseLogger *zap.Logger, config LoggerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := baseLogger

			if requestID, ok := r.Context().Value(config.RequestIDKey).(string); ok && requestID != "" {
				logger = logger.With(zap.String("request_id", requestID))
			}
			if userUUID, ok := r.Context().Value(config.UserUUIDKey).(string); ok && userUUID != "" {
				logger = logger.With(zap.String("user_uuid", userUUID))
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

// GetLogger retrieves the request-scoped logger from context
func GetLogger(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	return zap.NewNop()
}
