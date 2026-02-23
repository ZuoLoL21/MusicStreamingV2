package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// RequestIDConfig is the interface that services must implement to use RequestIDMiddleware (e.g. Config must have these methods)
type RequestIDConfig interface {
	GetRequestIDKey() any
}

func generateRequestID() string {
	return uuid.New().String()
}

func RequestIDMiddleware(config RequestIDConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			ctx := context.WithValue(r.Context(), config.GetRequestIDKey(), requestID)
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
