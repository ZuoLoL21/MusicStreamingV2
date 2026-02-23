package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// LoggingConfig is the interface that services must implement to use LoggingMiddleware (e.g. Config must have these methods)
type LoggingConfig interface {
	GetRequestIDKey() any
	GetUserUUIDKey() (any, bool) // Returns (key, hasUserUUID)
}

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func getIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	return r.RemoteAddr
}

func requestID(ctx context.Context, config LoggingConfig) string {
	id, _ := ctx.Value(config.GetRequestIDKey()).(string)
	return id
}

func userUUID(ctx context.Context, config LoggingConfig) (string, bool) {
	key, hasUserUUID := config.GetUserUUIDKey()
	if !hasUserUUID {
		return "", false
	}
	id, _ := ctx.Value(key).(string)
	return id, true
}

func LoggingMiddleware(logger *zap.Logger, config LoggingConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			defer func() {
				if rec := recover(); rec != nil {
					duration := time.Since(start)
					template, _ := mux.CurrentRoute(r).GetPathTemplate()
					ctx := r.Context()

					fields := []zap.Field{
						zap.Any("panic", rec),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String("route", template),
						zap.String("remote_addr", r.RemoteAddr),
						zap.Duration("duration", duration),
						zap.String("request_id", requestID(ctx, config)),
					}

					// Conditionally add user_uuid if the service tracks it
					if uuid, hasUUID := userUUID(ctx, config); hasUUID {
						fields = append(fields, zap.String("user_uuid", uuid))
					}

					logger.Error("panic recovered", fields...)
					http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			template, _ := mux.CurrentRoute(r).GetPathTemplate()

			ctx := r.Context()
			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("route", template),
				zap.Int("status", rw.status),
				zap.Int("bytes", rw.bytes),
				zap.Duration("duration", duration),
				zap.String("remote_addr", getIP(r)),
				zap.String("request_id", requestID(ctx, config)),
			}

			// Conditionally add user_uuid if the service tracks it
			if uuid, hasUUID := userUUID(ctx, config); hasUUID {
				fields = append(fields, zap.String("user_uuid", uuid))
			}

			switch {
			case rw.status >= 500:
				logger.Error("http request", fields...)
			case rw.status >= 400:
				logger.Warn("http request", fields...)
			default:
				logger.Info("http request", fields...)
			}
		})
	}
}
