package handlers

import (
	"backend/internal/consts"
	"net/http"
	"strconv"

	"backend/internal/di"

	sqlhandler "backend/sql/sqlc"
	libsdi "libs/di"

	"go.uber.org/zap"
)

type HistoryHandler struct {
	logger  *zap.Logger
	config  *di.Config
	returns *libsdi.ReturnManager
	db      consts.DB
}

func NewHistoryHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, db consts.DB) *HistoryHandler {
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

	limit, cursorTS, cursorID := parsePagination(r)
	history, err := h.db.GetListeningHistoryForUser(r.Context(), sqlhandler.GetListeningHistoryForUserParams{
		UserUuid: userUUID,
		Limit:    limit,
		Column3:  cursorTS,
		Uuid:     cursorID,
	})
	if err != nil {
		h.logger.Error("failed to get listening history", zap.Error(err))
		h.returns.ReturnError(w, "failed to get listening history", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, history, http.StatusOK)
}

func (h *HistoryHandler) GetTopMusicForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	limit, _, cursorID := parsePagination(r)
	var cursorCount interface{}
	if s := r.URL.Query().Get("cursor_count"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			cursorCount = n
		}
	}

	top, err := h.db.GetTopMusicForUser(r.Context(), sqlhandler.GetTopMusicForUserParams{
		UserUuid: userUUID,
		Limit:    limit,
		Column3:  cursorID,
		Column4:  cursorCount,
	})
	if err != nil {
		h.logger.Error("failed to get top music", zap.Error(err))
		h.returns.ReturnError(w, "failed to get top music", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, top, http.StatusOK)
}
