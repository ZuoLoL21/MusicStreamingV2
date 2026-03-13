package di

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ReturnManager handles HTTP response formatting and writing.
// It provides methods for returning JSON, text, error responses, and files
// with consistent formatting and logging.
type ReturnManager struct {
	logger *zap.Logger
}

// NewReturnManager creates a new ReturnManager with the given logger.
// The logger is used to log any errors that occur during response writing.
func NewReturnManager(logger *zap.Logger) *ReturnManager {
	return &ReturnManager{logger: logger}
}

// ReturnError writes an error response with the given message and HTTP status code.
//
// The error message is returned as JSON: {"error": "message"}
func (h *ReturnManager) ReturnError(w http.ResponseWriter, msg string, code int) {
	errResp := map[string]string{"error": msg}
	h.returnJSON(w, errResp, code)
}

// ReturnText writes a text message response with the given message and HTTP status code.
//
// The message is returned as JSON: {"message": "text"}
func (h *ReturnManager) ReturnText(w http.ResponseWriter, msg string, code int) {
	resp := map[string]string{"message": msg}
	h.returnJSON(w, resp, code)
}

// ReturnFile serves a file response with the given filename, modification time, and file content.
//
// It uses http.ServeContent to efficiently serve the file with support for range requests.
func (h *ReturnManager) ReturnFile(w http.ResponseWriter, r *http.Request, msg string, modTime time.Time, file io.ReadSeeker) {
	http.ServeContent(w, r, msg, modTime, file)
}

// ReturnJSON writes a JSON response with the given data and HTTP status code.
//
// The data is JSON-encoded and written to the response writer.
//
// Any encoding errors are logged but not returned to the client.
func (h *ReturnManager) ReturnJSON(w http.ResponseWriter, data interface{}, code int) {
	h.returnJSON(w, data, code)
}

// returnJSON is an internal helper that writes a JSON response with the given data and HTTP status code.
//
// It sets the Content-Type header to application/json, writes the status code,
// and encodes the data as JSON. Any encoding errors are logged but not returned to the client.
func (h *ReturnManager) returnJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Warn("failed to write response", zap.Error(err))
		}
	}
}
