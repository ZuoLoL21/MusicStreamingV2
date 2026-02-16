package middleware

import (
	"file-storage/internal/di"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

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

func LoggingMiddleware(logger *zap.Logger, config *di.Config) mux.MiddlewareFunc {
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
					logger.Error("panic recovered",
						zap.Any("panic", rec),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String("route", template),
						zap.String("remote_addr", r.RemoteAddr),
						zap.Duration("duration", duration),
						zap.String("request_id", ctx.Value(config.RequestIDKey).(string)),
					)
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
				zap.String("request_id", ctx.Value(config.RequestIDKey).(string)),
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
