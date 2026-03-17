package handlers

import (
	"backend/internal/consts"
	"net/http"

	"backend/internal/di"
	"backend/internal/storage"
	sqlhandler "backend/sql/sqlc"

	libsdi "libs/di"
	libsmiddleware "libs/middleware"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type AlbumHandler struct {
	logger      *zap.Logger
	config      *di.Config
	returns     *libsdi.ReturnManager
	db          consts.DB
	fileStorage storage.FileStorageClient
}

func NewAlbumHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, db consts.DB, fileStorage storage.FileStorageClient) *AlbumHandler {
	return &AlbumHandler{
		logger:      logger,
		config:      config,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
	}
}

// checkAlbumAccess parses the album UUID from the route, fetches the album to
// resolve its artist, and verifies the calling user has at least the given role.
func (h *AlbumHandler) checkAlbumAccess(w http.ResponseWriter, r *http.Request, role sqlhandler.ArtistMemberRole) (albumUUID pgtype.UUID, ok bool) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	albumUUID, ok = parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	album, err := h.db.GetAlbum(r.Context(), albumUUID)
	if handleDBError(w, err, "album not found", h.logger, h.returns) {
		ok = false
		return
	}

	if !checkArtistRole(r.Context(), h.db, album.FromArtist, userUUID, role) {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		ok = false
	}

	return
}

func (h *AlbumHandler) GetAlbum(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	albumUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	album, err := h.db.GetAlbum(r.Context(), albumUUID)
	if handleDBError(w, err, "album not found", logger, h.returns) {
		return
	}

	applyDefaultImageIfEmpty(&album.ImagePath, "album")
	h.returns.ReturnJSON(w, album, http.StatusOK)
}

func (h *AlbumHandler) GetAlbumsForArtist(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	artistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	limit, cursorTS, cursorID := parsePagination(r)
	albums, err := h.db.GetAlbumsForArtist(r.Context(), sqlhandler.GetAlbumsForArtistParams{
		FromArtist: artistUUID,
		Limit:      limit,
		Column3:    cursorTS,
		Uuid:       cursorID,
	})
	if err != nil {
		logger.Warn("failed to get albums for artist",
			zap.String("artist_uuid", uuidToString(artistUUID)),
			zap.Error(err))
		h.returns.ReturnError(w, "failed to get albums", http.StatusInternalServerError)
		return
	}

	for i := range albums {
		applyDefaultImageIfEmpty(&albums[i].ImagePath, "album")
	}

	logger.Debug("albums retrieved successfully",
		zap.String("artist_uuid", uuidToString(artistUUID)),
		zap.Int("count", len(albums)))
	h.returns.ReturnJSON(w, albums, http.StatusOK)
}

func (h *AlbumHandler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns, h.logger) {
		return
	}

	artistUUIDStr := r.FormValue("artist_uuid")
	if artistUUIDStr == "" {
		h.returns.ReturnError(w, "artist_uuid required", http.StatusBadRequest)
		return
	}
	rawOriginalName := r.FormValue("original_name")

	originalName, err := validateStringField(rawOriginalName, "original_name", 1, 200)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var description *string
	if descVal := r.FormValue("description"); descVal != "" {
		description = &descVal
	}

	// Generate album ID
	albumID := uuid.New().String()

	// Optional image upload
	imagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		consts.PicturesAlbumFolder, albumID, "image", h.logger, h.returns)
	if !ok {
		return
	}

	artistUUID, err := uuidToPgtype(artistUUIDStr)
	if err != nil {
		h.returns.ReturnError(w, "invalid artist_uuid", http.StatusBadRequest)
		return
	}

	if !checkArtistRole(r.Context(), h.db, artistUUID, userUUID, sqlhandler.ArtistMemberRoleMember) {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		return
	}

	descText := optionalStringToPgtype(description)

	if err := h.db.CreateAlbum(r.Context(), sqlhandler.CreateAlbumParams{
		FromArtist:   artistUUID,
		OriginalName: originalName,
		Description:  descText,
		ImagePath:    imagePath,
	}); err != nil {
		logger.Error("failed to create album", zap.Error(err))

		if imagePath.Valid {
			cleanupImage(r.Context(), h.fileStorage, consts.PicturesAlbumFolder, albumID, h.logger)
		}
		h.returns.ReturnError(w, "failed to create album", http.StatusInternalServerError)
		return
	}

	logger.Info("album created successfully",
		zap.String("album_uuid", albumID),
		zap.String("original_name", originalName))
	h.returns.ReturnText(w, "album created", http.StatusCreated)
}

type updateAlbumRequest struct {
	OriginalName string  `json:"original_name" validate:"required,max=255"`
	Description  *string `json:"description"`
}

func (h *AlbumHandler) UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	albumUUID, ok := h.checkAlbumAccess(w, r, sqlhandler.ArtistMemberRoleManager)
	if !ok {
		return
	}

	body, ok := decodeBody[updateAlbumRequest](w, r, h.returns)
	if !ok {
		return
	}

	originalName, err := validateStringField(body.OriginalName, "original_name", 1, 200)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var description pgtype.Text
	if body.Description != nil {
		description = pgtype.Text{String: *body.Description, Valid: true}
	}

	if err := h.db.UpdateAlbum(r.Context(), sqlhandler.UpdateAlbumParams{
		Uuid:         albumUUID,
		OriginalName: originalName,
		Description:  description,
	}); err != nil {
		logger.Error("failed to update album", zap.Error(err))
		h.returns.ReturnError(w, "failed to update album", http.StatusInternalServerError)
		return
	}

	logger.Info("album updated successfully",
		zap.String("album_uuid", uuidToString(albumUUID)))
	h.returns.ReturnText(w, "album updated", http.StatusOK)
}

func (h *AlbumHandler) UpdateAlbumImage(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	albumUUID, ok := h.checkAlbumAccess(w, r, sqlhandler.ArtistMemberRoleManager)
	if !ok {
		return
	}

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns, h.logger) {
		return
	}

	// Album ID
	imageID := uuid.UUID(albumUUID.Bytes).String()

	// Update
	imagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		consts.PicturesAlbumFolder, imageID, "image", h.logger, h.returns)
	if !ok {
		return
	}

	if !imagePath.Valid {
		h.returns.ReturnError(w, "image file required", http.StatusBadRequest)
		return
	}

	// Update database
	if err := h.db.UpdateAlbumImage(r.Context(), sqlhandler.UpdateAlbumImageParams{
		Uuid:      albumUUID,
		ImagePath: imagePath,
	}); err != nil {
		logger.Error("failed to update album image in database", zap.Error(err))
		h.returns.ReturnError(w, "failed to update album image", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "album image updated", http.StatusOK)
}

func (h *AlbumHandler) DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	albumUUID, ok := h.checkAlbumAccess(w, r, sqlhandler.ArtistMemberRoleOwner)
	if !ok {
		return
	}

	if err := h.db.DeleteAlbum(r.Context(), albumUUID); err != nil {
		logger.Error("failed to delete album", zap.Error(err))
		h.returns.ReturnError(w, "failed to delete album", http.StatusInternalServerError)
		return
	}

	logger.Info("album deleted successfully",
		zap.String("album_uuid", uuidToString(albumUUID)))
	h.returns.ReturnText(w, "album deleted", http.StatusOK)
}
