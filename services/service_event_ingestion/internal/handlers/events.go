package handlers

import (
	"encoding/json"
	"event_ingestion/internal/di"
	"net/http"
	"strings"

	libsdi "libs/di"

	"github.com/google/uuid"

	"go.uber.org/zap"
)

type EventHandler struct {
	logger     *zap.Logger
	config     *di.Config
	returns    *libsdi.ReturnManager
	clickhouse *di.ClickHouseClient
}

func NewEventHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, clickhouse *di.ClickHouseClient) *EventHandler {
	return &EventHandler{
		logger:     logger,
		config:     config,
		returns:    returns,
		clickhouse: clickhouse,
	}
}

func (h *EventHandler) IngestListenEvent(w http.ResponseWriter, r *http.Request) {
	var req di.ListenEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid listen event request", zap.Error(err))
		h.returns.ReturnError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if req.UserUUID == uuid.Nil || req.MusicUUID == uuid.Nil || req.ArtistUUID == uuid.Nil {
		h.logger.Warn("missing required UUIDs in listen event",
			zap.String("user_uuid", req.UserUUID.String()),
			zap.String("music_uuid", req.MusicUUID.String()),
			zap.String("artist_uuid", req.ArtistUUID.String()),
		)
		h.returns.ReturnError(w, "Missing required UUIDs", http.StatusBadRequest)
		return
	}
	if req.CompletionRatio < 0 || req.CompletionRatio > 1 {
		h.returns.ReturnError(w, "Completion ratio must be between 0 and 1", http.StatusBadRequest)
		return
	}

	// Insert into ClickHouse
	if err := h.clickhouse.InsertListenEvent(r.Context(), req); err != nil {
		h.logger.Error("failed to insert listen event", zap.Error(err))
		h.returns.ReturnError(w, "Failed to ingest event", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, map[string]interface{}{
		"success": true,
	}, http.StatusOK)
}

func (h *EventHandler) IngestLikeEvent(w http.ResponseWriter, r *http.Request) {
	var req di.LikeEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid like event request", zap.Error(err))
		h.returns.ReturnError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if req.UserUUID == uuid.Nil || req.MusicUUID == uuid.Nil || req.ArtistUUID == uuid.Nil {
		h.returns.ReturnError(w, "Missing required UUIDs", http.StatusBadRequest)
		return
	}

	// Insert into ClickHouse
	if err := h.clickhouse.InsertLikeEvent(r.Context(), req); err != nil {
		h.logger.Error("failed to insert like event", zap.Error(err))
		h.returns.ReturnError(w, "Failed to ingest event", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, map[string]interface{}{
		"success": true,
	}, http.StatusOK)
}

func (h *EventHandler) IngestThemeEvent(w http.ResponseWriter, r *http.Request) {
	var req di.ThemeEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid theme event request", zap.Error(err))
		h.returns.ReturnError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if req.MusicUUID == uuid.Nil || req.Theme == "" {
		h.returns.ReturnError(w, "Missing music_uuid or theme", http.StatusBadRequest)
		return
	}

	// Upsert theme into ClickHouse
	if err := h.clickhouse.UpsertTheme(r.Context(), req); err != nil {
		h.logger.Error("failed to upsert theme", zap.Error(err))
		h.returns.ReturnError(w, "Failed to ingest event", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, map[string]interface{}{
		"success": true,
	}, http.StatusOK)
}

func (h *EventHandler) IngestUserDimEvent(w http.ResponseWriter, r *http.Request) {
	var req di.UserDimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid user dim request", zap.Error(err))
		h.returns.ReturnError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if req.UserUUID == uuid.Nil || req.Country == "" {
		h.returns.ReturnError(w, "Missing user_uuid or country", http.StatusBadRequest)
		return
	}
	if len(req.Country) != 2 {
		h.returns.ReturnError(w, "Country must be ISO 3166-1 alpha-2 code (2 chars)", http.StatusBadRequest)
		return
	}
	if req.Country != strings.ToUpper(req.Country) {
		h.returns.ReturnError(w, "Country must be uppercase ISO 3166-1 alpha-2 code", http.StatusBadRequest)
		return
	}

	// Upsert user dimension into ClickHouse
	if err := h.clickhouse.UpsertUserDim(r.Context(), req); err != nil {
		h.logger.Error("failed to upsert user dim", zap.Error(err))
		h.returns.ReturnError(w, "Failed to ingest event", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, map[string]interface{}{
		"success": true,
	}, http.StatusOK)
}
