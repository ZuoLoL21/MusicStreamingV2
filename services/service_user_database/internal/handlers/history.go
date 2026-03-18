package handlers

import (
	"backend/internal/consts"
	"net/http"
	"strconv"

	"backend/internal/di"

	sqlhandler "backend/sql/sqlc"
	libsdi "libs/di"
	libsmiddleware "libs/middleware"

	"go.uber.org/zap"
)

type HistoryHandler struct {
	config  *di.Config
	returns *libsdi.ReturnManager
	db      consts.DB
}

func NewHistoryHandler(config *di.Config, returns *libsdi.ReturnManager, db consts.DB) *HistoryHandler {
	return &HistoryHandler{
		config:  config,
		returns: returns,
		db:      db,
	}
}

func (h *HistoryHandler) GetListeningHistoryForUser(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

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
		logger.Warn("failed to get listening history",
			zap.String("user_uuid", uuidToString(userUUID)),
			zap.Error(err))
		h.returns.ReturnError(w, "failed to get listening history", http.StatusInternalServerError)
		return
	}

	logger.Debug("listening history retrieved successfully",
		zap.String("user_uuid", uuidToString(userUUID)),
		zap.Int("count", len(history)))
	h.returns.ReturnJSON(w, history, http.StatusOK)
}

func (h *HistoryHandler) GetTopMusicForUser(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

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
		logger.Warn("failed to get top music",
			zap.String("user_uuid", uuidToString(userUUID)),
			zap.Error(err))
		h.returns.ReturnError(w, "failed to get top music", http.StatusInternalServerError)
		return
	}

	logger.Debug("top music retrieved successfully",
		zap.String("user_uuid", uuidToString(userUUID)),
		zap.Int("count", len(top)))
	h.returns.ReturnJSON(w, top, http.StatusOK)
}
