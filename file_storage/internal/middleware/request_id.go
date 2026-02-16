package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func generateRequestID() string {
	return uuid.New().String()
}

func RequestIDMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
