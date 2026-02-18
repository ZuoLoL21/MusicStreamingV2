package handlers

import (
	"net/http"
	"strconv"

	"backend/internal/di"
	sqlhandler "backend/sql/sqlc"

	"go.uber.org/zap"
)

type HistoryHandler struct {
	logger  *zap.Logger
	config  *di.Config
	returns *di.ReturnManager
	db      *sqlhandler.Queries
}

func NewHistoryHandler(logger *zap.Logger, config *di.Config, returns *di.ReturnManager, db *sqlhandler.Queries) *HistoryHandler {
	return &HistoryHandler{
		logger:  logger,
		config:  config,
		returns: returns,
		db:      db,
	}
}

func (h *HistoryHandler) GetListeningHistoryForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	history, err := h.db.GetListeningHistoryForUser(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("failed to get listening history", zap.Error(err))
		h.returns.ReturnError(w, "failed to get listening history", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, history, http.StatusOK)
}

func (h *HistoryHandler) GetRecentlyPlayedForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	limit := int32(10)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = int32(parsed)
		}
	}

	history, err := h.db.GetRecentlyPlayedForUser(r.Context(), sqlhandler.GetRecentlyPlayedForUserParams{
		UserUuid: userUUID,
		Limit:    limit,
	})
	if err != nil {
		h.logger.Error("failed to get recently played", zap.Error(err))
		h.returns.ReturnError(w, "failed to get recently played", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, history, http.StatusOK)
}

func (h *HistoryHandler) GetTopMusicForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	limit := int32(10)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = int32(parsed)
		}
	}

	top, err := h.db.GetTopMusicForUser(r.Context(), sqlhandler.GetTopMusicForUserParams{
		UserUuid: userUUID,
		Limit:    limit,
	})
	if err != nil {
		h.logger.Error("failed to get top music", zap.Error(err))
		h.returns.ReturnError(w, "failed to get top music", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, top, http.StatusOK)
}
