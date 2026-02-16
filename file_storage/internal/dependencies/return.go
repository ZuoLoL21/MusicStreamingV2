package dependencies

import (
	"encoding/json"
	"net/http"

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
	h.ReturnJSON(w, errResp, code)
}

func (h *ReturnManager) ReturnJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Warn("failed to write response", zap.Error(err))
		}
	}
}
