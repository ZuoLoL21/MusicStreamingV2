package handlers

import (
	"backend/internal/consts"
	"net/http"

	"backend/internal/di"
	"backend/internal/storage"

	sqlhandler "backend/sql/sqlc"
	libsdi "libs/di"
	libsmiddleware "libs/middleware"

	"go.uber.org/zap"
)

type LikesHandler struct {
	logger      *zap.Logger
	config      *di.Config
	returns     *libsdi.ReturnManager
	db          consts.DB
	fileStorage storage.FileStorageClient
}

func NewLikesHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, db consts.DB, fileStorage storage.FileStorageClient) *LikesHandler {
	return &LikesHandler{
		logger:      logger,
		config:      config,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
	}
}

func (h *LikesHandler) GetLikesForUser(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

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
		logger.Warn("failed to get likes for user",
			zap.String("user_uuid", uuidToString(userUUID)),
			zap.Error(err))
		h.returns.ReturnError(w, "failed to get likes", http.StatusInternalServerError)
		return
	}

	for i := range likes {
		likes[i].PathInFileStorage = convertPathToFileURL(likes[i].PathInFileStorage)
		applyDefaultImageIfEmpty(&likes[i].ImagePath, "music")
	}

	logger.Debug("likes retrieved successfully",
		zap.String("user_uuid", uuidToString(userUUID)),
		zap.Int("count", len(likes)))
	h.returns.ReturnJSON(w, likes, http.StatusOK)
}

func (h *LikesHandler) IsLiked(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

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
		logger.Warn("failed to check like status",
			zap.String("user_uuid", uuidToString(userUUID)),
			zap.String("music_uuid", uuidToString(musicUUID)),
			zap.Error(err))
		h.returns.ReturnError(w, "failed to check like status", http.StatusInternalServerError)
		return
	}

	logger.Debug("like status checked",
		zap.String("user_uuid", uuidToString(userUUID)),
		zap.String("music_uuid", uuidToString(musicUUID)),
		zap.Bool("liked", liked))
	h.returns.ReturnJSON(w, map[string]bool{"liked": liked}, http.StatusOK)
}

func (h *LikesHandler) LikeMusic(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

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
		logger.Error("failed to like music", zap.Error(err))
		h.returns.ReturnError(w, "failed to like music", http.StatusInternalServerError)
		return
	}

	logger.Info("music liked successfully",
		zap.String("user_uuid", uuidToString(userUUID)),
		zap.String("music_uuid", uuidToString(musicUUID)))
	h.returns.ReturnText(w, "music liked", http.StatusOK)
}

func (h *LikesHandler) UnlikeMusic(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

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
		logger.Error("failed to unlike music", zap.Error(err))
		h.returns.ReturnError(w, "failed to unlike music", http.StatusInternalServerError)
		return
	}

	logger.Info("music unliked successfully",
		zap.String("user_uuid", uuidToString(userUUID)),
		zap.String("music_uuid", uuidToString(musicUUID)))
	h.returns.ReturnText(w, "music unliked", http.StatusOK)
}
