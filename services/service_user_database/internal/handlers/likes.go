package handlers

import (
	"net/http"

	"backend/internal/di"
	"backend/internal/storage"

	sqlhandler "backend/sql/sqlc"
	libsdi "libs/di"

	"go.uber.org/zap"
)

type LikesHandler struct {
	logger      *zap.Logger
	config      *di.Config
	returns     *libsdi.ReturnManager
	db          DB
	fileStorage storage.FileStorageClient
}

func NewLikesHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, db DB, fileStorage storage.FileStorageClient) *LikesHandler {
	return &LikesHandler{
		logger:      logger,
		config:      config,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
	}
}

func (h *LikesHandler) GetLikesForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	limit, cursorTS, cursorID := parsePagination(r)
	likes, err := h.db.GetLikesForUser(r.Context(), sqlhandler.GetLikesForUserParams{
		FromUser: userUUID,
		Limit:    limit,
		Column3:  cursorTS,
		Uuid:     cursorID,
	})
	if err != nil {
		h.logger.Error("failed to get likes for user", zap.Error(err))
		h.returns.ReturnError(w, "failed to get likes", http.StatusInternalServerError)
		return
	}

	for i := range likes {
		likes[i].PathInFileStorage = h.fileStorage.BuildPublicURL(likes[i].PathInFileStorage)
		applyDefaultImageIfEmpty(&likes[i].ImagePath, h.fileStorage, "music")
	}

	h.returns.ReturnJSON(w, likes, http.StatusOK)
}

func (h *LikesHandler) IsLiked(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	musicUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	liked, err := h.db.IsLiked(r.Context(), sqlhandler.IsLikedParams{
		FromUser: userUUID,
		ToMusic:  musicUUID,
	})
	if err != nil {
		h.logger.Error("failed to check like status", zap.Error(err))
		h.returns.ReturnError(w, "failed to check like status", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, map[string]bool{"liked": liked}, http.StatusOK)
}

func (h *LikesHandler) LikeMusic(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	musicUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	if err := h.db.LikeMusic(r.Context(), sqlhandler.LikeMusicParams{
		FromUser: userUUID,
		ToMusic:  musicUUID,
	}); err != nil {
		h.logger.Error("failed to like music", zap.Error(err))
		h.returns.ReturnError(w, "failed to like music", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "music liked", http.StatusOK)
}

func (h *LikesHandler) UnlikeMusic(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	musicUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	if err := h.db.UnlikeMusic(r.Context(), sqlhandler.UnlikeMusicParams{
		FromUser: userUUID,
		ToMusic:  musicUUID,
	}); err != nil {
		h.logger.Error("failed to unlike music", zap.Error(err))
		h.returns.ReturnError(w, "failed to unlike music", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "music unliked", http.StatusOK)
}
