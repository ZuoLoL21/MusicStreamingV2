package handlers

import (
	"net/http"

	"backend/internal/di"

	sqlhandler "backend/sql/sqlc"
	libsdi "libs/di"

	"go.uber.org/zap"
)

type FollowsHandler struct {
	logger  *zap.Logger
	config  *di.Config
	returns *libsdi.ReturnManager
	db      DB
}

func NewFollowsHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, db DB) *FollowsHandler {
	return &FollowsHandler{
		logger:  logger,
		config:  config,
		returns: returns,
		db:      db,
	}
}

func (h *FollowsHandler) GetFollowersForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	limit, cursorTS, cursorID := parsePagination(r)
	followers, err := h.db.GetFollowersForUser(r.Context(), sqlhandler.GetFollowersForUserParams{
		ToUser:  userUUID,
		Limit:   limit,
		Column3: cursorTS,
		Uuid:    cursorID,
	})
	if err != nil {
		h.logger.Error("failed to get followers for user", zap.Error(err))
		h.returns.ReturnError(w, "failed to get followers", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, followers, http.StatusOK)
}

func (h *FollowsHandler) GetFollowingUsersForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	limit, cursorTS, cursorID := parsePagination(r)
	following, err := h.db.GetFollowedUsersForUser(r.Context(), sqlhandler.GetFollowedUsersForUserParams{
		FromUser: userUUID,
		Limit:    limit,
		Column3:  cursorTS,
		Uuid:     cursorID,
	})
	if err != nil {
		h.logger.Error("failed to get following users", zap.Error(err))
		h.returns.ReturnError(w, "failed to get following users", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, following, http.StatusOK)
}

func (h *FollowsHandler) GetFollowersForArtist(w http.ResponseWriter, r *http.Request) {
	artistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	limit, cursorTS, cursorID := parsePagination(r)
	followers, err := h.db.GetFollowersForArtist(r.Context(), sqlhandler.GetFollowersForArtistParams{
		ToArtist: artistUUID,
		Limit:    limit,
		Column3:  cursorTS,
		Uuid:     cursorID,
	})
	if err != nil {
		h.logger.Error("failed to get followers for artist", zap.Error(err))
		h.returns.ReturnError(w, "failed to get followers", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, followers, http.StatusOK)
}

func (h *FollowsHandler) GetFollowedArtistsForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	limit, cursorTS, cursorID := parsePagination(r)
	artists, err := h.db.GetFollowedArtistsForUser(r.Context(), sqlhandler.GetFollowedArtistsForUserParams{
		FromUser: userUUID,
		Limit:    limit,
		Column3:  cursorTS,
		Uuid:     cursorID,
	})
	if err != nil {
		h.logger.Error("failed to get followed artists", zap.Error(err))
		h.returns.ReturnError(w, "failed to get followed artists", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, artists, http.StatusOK)
}

func (h *FollowsHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	fromUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	toUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	if err := h.db.FollowUser(r.Context(), sqlhandler.FollowUserParams{
		FromUser: fromUUID,
		ToUser:   toUUID,
	}); err != nil {
		h.logger.Error("failed to follow user", zap.Error(err))
		h.returns.ReturnError(w, "failed to follow user", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "user followed", http.StatusOK)
}

func (h *FollowsHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	fromUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	toUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	if err := h.db.UnfollowUser(r.Context(), sqlhandler.UnfollowUserParams{
		FromUser: fromUUID,
		ToUser:   toUUID,
	}); err != nil {
		h.logger.Error("failed to unfollow user", zap.Error(err))
		h.returns.ReturnError(w, "failed to unfollow user", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "user unfollowed", http.StatusOK)
}

func (h *FollowsHandler) FollowArtist(w http.ResponseWriter, r *http.Request) {
	fromUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	artistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	if err := h.db.FollowArtist(r.Context(), sqlhandler.FollowArtistParams{
		FromUser: fromUUID,
		ToArtist: artistUUID,
	}); err != nil {
		h.logger.Error("failed to follow artist", zap.Error(err))
		h.returns.ReturnError(w, "failed to follow artist", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "artist followed", http.StatusOK)
}

func (h *FollowsHandler) UnfollowArtist(w http.ResponseWriter, r *http.Request) {
	fromUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	artistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	if err := h.db.UnfollowArtist(r.Context(), sqlhandler.UnfollowArtistParams{
		FromUser: fromUUID,
		ToArtist: artistUUID,
	}); err != nil {
		h.logger.Error("failed to unfollow artist", zap.Error(err))
		h.returns.ReturnError(w, "failed to unfollow artist", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "artist unfollowed", http.StatusOK)
}
