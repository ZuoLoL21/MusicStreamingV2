package middleware

import (
	"context"
	"net/http"
	"popularity/internal/di"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func generateRequestID() string {
	return uuid.New().String()
}

func RequestIDMiddleware(config *di.Config) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			ctx := context.WithValue(r.Context(), config.RequestIDKey, requestID)
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
