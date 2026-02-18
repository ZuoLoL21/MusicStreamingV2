package handlers

import (
	"net/http"

	"backend/internal/di"
	sql_handler "backend/sql/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type AlbumHandler struct {
	logger  *zap.Logger
	config  *di.Config
	returns *di.ReturnManager
	db      *sql_handler.Queries
}

func NewAlbumHandler(logger *zap.Logger, config *di.Config, returns *di.ReturnManager, db *sql_handler.Queries) *AlbumHandler {
	return &AlbumHandler{
		logger:  logger,
		config:  config,
		returns: returns,
		db:      db,
	}
}

func (h *AlbumHandler) GetAlbum(w http.ResponseWriter, r *http.Request) {
	albumUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	album, err := h.db.GetAlbum(r.Context(), albumUUID)
	if err != nil {
		h.returns.ReturnError(w, "album not found", http.StatusNotFound)
		return
	}

	h.returns.ReturnJSON(w, album, http.StatusOK)
}

func (h *AlbumHandler) GetAlbumsForArtist(w http.ResponseWriter, r *http.Request) {
	artistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	albums, err := h.db.GetAlbumsForArtist(r.Context(), artistUUID)
	if err != nil {
		h.logger.Error("failed to get albums for artist", zap.Error(err))
		h.returns.ReturnError(w, "failed to get albums", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, albums, http.StatusOK)
}

type createAlbumRequest struct {
	ArtistUUID   string  `json:"artist_uuid" validate:"required"`
	OriginalName string  `json:"original_name" validate:"required,max=255"`
	Description  *string `json:"description"`
	ImagePath    *string `json:"image_path"`
}

func (h *AlbumHandler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	body, ok := decodeBody[createAlbumRequest](w, r, h.returns)
	if !ok {
		return
	}

	artistUUID, err := uuidToPgtype(body.ArtistUUID)
	if err != nil {
		h.returns.ReturnError(w, "invalid artist_uuid", http.StatusBadRequest)
		return
	}

	if !checkArtistRole(r.Context(), h.db, artistUUID, userUUID, sql_handler.ArtistMemberRoleMember) {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		return
	}

	var description pgtype.Text
	if body.Description != nil {
		description = pgtype.Text{String: *body.Description, Valid: true}
	}

	var imagePath pgtype.Text
	if body.ImagePath != nil {
		imagePath = pgtype.Text{String: *body.ImagePath, Valid: true}
	}

	if err := h.db.CreateAlbum(r.Context(), sql_handler.CreateAlbumParams{
		FromArtist:   artistUUID,
		OriginalName: body.OriginalName,
		Description:  description,
		ImagePath:    imagePath,
	}); err != nil {
		h.logger.Error("failed to create album", zap.Error(err))
		h.returns.ReturnError(w, "failed to create album", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "album created", http.StatusCreated)
}

type updateAlbumRequest struct {
	OriginalName string  `json:"original_name" validate:"required,max=255"`
	Description  *string `json:"description"`
}

func (h *AlbumHandler) UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	albumUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	album, err := h.db.GetAlbum(r.Context(), albumUUID)
	if err != nil {
		h.returns.ReturnError(w, "album not found", http.StatusNotFound)
		return
	}

	if !checkArtistRole(r.Context(), h.db, album.FromArtist, userUUID, sql_handler.ArtistMemberRoleManager) {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		return
	}

	body, ok := decodeBody[updateAlbumRequest](w, r, h.returns)
	if !ok {
		return
	}

	var description pgtype.Text
	if body.Description != nil {
		description = pgtype.Text{String: *body.Description, Valid: true}
	}

	if err := h.db.UpdateAlbum(r.Context(), sql_handler.UpdateAlbumParams{
		Uuid:         albumUUID,
		OriginalName: body.OriginalName,
		Description:  description,
	}); err != nil {
		h.logger.Error("failed to update album", zap.Error(err))
		h.returns.ReturnError(w, "failed to update album", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "album updated", http.StatusOK)
}

type updateAlbumImageRequest struct {
	ImagePath string `json:"image_path" validate:"required"`
}

func (h *AlbumHandler) UpdateAlbumImage(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	albumUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	album, err := h.db.GetAlbum(r.Context(), albumUUID)
	if err != nil {
		h.returns.ReturnError(w, "album not found", http.StatusNotFound)
		return
	}

	if !checkArtistRole(r.Context(), h.db, album.FromArtist, userUUID, sql_handler.ArtistMemberRoleManager) {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		return
	}

	body, ok := decodeBody[updateAlbumImageRequest](w, r, h.returns)
	if !ok {
		return
	}

	if err := h.db.UpdateAlbumImage(r.Context(), sql_handler.UpdateAlbumImageParams{
		Uuid:      albumUUID,
		ImagePath: pgtype.Text{String: body.ImagePath, Valid: true},
	}); err != nil {
		h.logger.Error("failed to update album image", zap.Error(err))
		h.returns.ReturnError(w, "failed to update album image", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "album image updated", http.StatusOK)
}

func (h *AlbumHandler) DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	albumUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	album, err := h.db.GetAlbum(r.Context(), albumUUID)
	if err != nil {
		h.returns.ReturnError(w, "album not found", http.StatusNotFound)
		return
	}

	if !checkArtistRole(r.Context(), h.db, album.FromArtist, userUUID, sql_handler.ArtistMemberRoleOwner) {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := h.db.DeleteAlbum(r.Context(), albumUUID); err != nil {
		h.logger.Error("failed to delete album", zap.Error(err))
		h.returns.ReturnError(w, "failed to delete album", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "album deleted", http.StatusOK)
}
