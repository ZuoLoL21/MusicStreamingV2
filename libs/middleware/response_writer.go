package middleware

import (
	"net/http"
)

// TrackedResponseWriter wraps http.ResponseWriter to capture response metadata.
// It tracks the HTTP status code and the number of bytes written to the response body.
//
// This wrapper is used by multiple middleware (metrics, failure_recovery, etc.)
// to collect information about the HTTP response that would otherwise not be
// accessible after the response has been written.
type TrackedResponseWriter struct {
	http.ResponseWriter
	StatusCode   int
	BytesWritten int64
}

func (rw *TrackedResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *TrackedResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.BytesWritten += int64(n)
	return n, err
}

func WrapResponseWriter(w http.ResponseWriter) *TrackedResponseWriter {
	if tracked, ok := w.(*TrackedResponseWriter); ok {
		return tracked
	}

	return &TrackedResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}
