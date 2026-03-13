package middleware

import (
	"context"
	"libs/helpers"
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

func requestID(ctx context.Context) string {
	return helpers.GetRequestIDFromContext(ctx)
}

func userUUID(ctx context.Context) string {
	return helpers.GetUserUUIDFromContext(ctx)
}

// FailureRecoveryMiddleware is a middleware that wraps an HTTP handler to provide
// panic recovery and request logging. It recovers from any panics that occur during
// request processing, logs the panic with relevant request details, and returns a
// 500 Internal Server Error response.
//
// Additionally, it logs all HTTP requests with
// method, path, route, status code, response size, duration, and client information.
func FailureRecoveryMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			defer func() {
				rec := recover()
				if rec != nil {
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
						zap.String("request_id", requestID(ctx)),
					}

					if uuid := userUUID(ctx); uuid != "" {
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
				zap.String("request_id", requestID(ctx)),
			}

			if uuid := userUUID(ctx); uuid != "" {
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
