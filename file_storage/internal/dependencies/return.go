package dependencies

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ReturnManager struct {
	logger *zap.Logger
	config *Config
}

func GetReturnManager(logger *zap.Logger, config *Config) *ReturnManager {
	return &ReturnManager{logger: logger, config: config}
}

func (h *ReturnManager) ReturnError(w http.ResponseWriter, r *http.Request, msg string, code int) {
	if code >= 500 {
		h.logger.Error("server error",
			zap.String("msg", msg),
			zap.Int("code", code),
			zap.String("request_id", r.Context().Value(h.config.RequestIDKey).(string)),
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()))
	} else if code == 400 {
		h.logger.Info("bad request",
			zap.String("msg", msg),
			zap.Int("code", code),
			zap.String("request_id", r.Context().Value(h.config.RequestIDKey).(string)),
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()))
	} else if code > 400 {
		h.logger.Warn("problem with request",
			zap.String("msg", msg),
			zap.Int("code", code),
			zap.String("request_id", r.Context().Value(h.config.RequestIDKey).(string)),
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()))
	}

	errResp := map[string]string{"error": msg}
	h.returnJSON(w, errResp, code)
}

func (h *ReturnManager) ReturnText(w http.ResponseWriter, msg string, code int) {
	resp := map[string]string{"message": msg}
	h.returnJSON(w, resp, code)
}

func (h *ReturnManager) ReturnFile(w http.ResponseWriter, r *http.Request, msg string, modtime time.Time, file io.ReadSeeker) {
	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, msg, modtime, file)
}

func (h *ReturnManager) ReturnJSON(w http.ResponseWriter, data interface{}, code int) {
	h.returnJSON(w, data, code)
}

func (h *ReturnManager) returnJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Warn("failed to write response", zap.Error(err))
		}
	}
}
