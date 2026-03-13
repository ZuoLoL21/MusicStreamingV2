package middleware

import (
	"context"
	"libs/consts"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func generateRequestID() string {
	return uuid.New().String()
}

// RequestIDMiddleware returns a Gorilla Mux middleware that handles request ID tracking.
//
// It extracts the request ID from the X-Request-ID header if present, otherwise generates
// a new UUID. The request ID is stored in the request context and also set as a response header.
//
// This enables request tracing across services.
func RequestIDMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			ctx := context.WithValue(r.Context(), consts.RequestIDKey, requestID)
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
