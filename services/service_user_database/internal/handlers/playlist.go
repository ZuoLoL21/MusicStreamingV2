package handlers

import (
	"backend/internal/consts"
	"net/http"
	"strconv"

	"backend/internal/di"
	"backend/internal/storage"

	sqlhandler "backend/sql/sqlc"
	libsdi "libs/di"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type PlaylistHandler struct {
	logger      *zap.Logger
	config      *di.Config
	returns     *libsdi.ReturnManager
	db          consts.DB
	fileStorage storage.FileStorageClient
}

func NewPlaylistHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, db consts.DB, fileStorage storage.FileStorageClient) *PlaylistHandler {
	return &PlaylistHandler{
		logger:      logger,
		config:      config,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
	}
}

// checkPlaylistOwnership parses the playlist UUID from the route, fetches the
// playlist, and verifies the calling user is its owner.
func (h *PlaylistHandler) checkPlaylistOwnership(w http.ResponseWriter, r *http.Request) (userUUID, playlistUUID pgtype.UUID, ok bool) {
	userUUID, ok = userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	playlistUUID, ok = parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	playlist, err := h.db.GetPlaylist(r.Context(), playlistUUID)
	if err != nil {
		h.returns.ReturnError(w, "playlist not found", http.StatusNotFound)
		ok = false
		return
	}

	if playlist.FromUser.Bytes != userUUID.Bytes {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		ok = false
	}

	return
}

func (h *PlaylistHandler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	playlist, err := h.db.GetPlaylist(r.Context(), playlistUUID)
	if err != nil {
		h.returns.ReturnError(w, "playlist not found", http.StatusNotFound)
		return
	}

	applyDefaultImageIfEmpty(&playlist.ImagePath, "playlist")
	h.returns.ReturnJSON(w, playlist, http.StatusOK)
}

func (h *PlaylistHandler) GetPlaylistsForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	limit, cursorTS, cursorID := parsePagination(r)
	playlists, err := h.db.GetPlaylistsForUser(r.Context(), sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    limit,
		Column3:  cursorTS,
		Uuid:     cursorID,
	})
	if err != nil {
		h.logger.Error("failed to get playlists for user", zap.Error(err))
		h.returns.ReturnError(w, "failed to get playlists", http.StatusInternalServerError)
		return
	}

	for i := range playlists {
		applyDefaultImageIfEmpty(&playlists[i].ImagePath, "playlist")
	}

	h.returns.ReturnJSON(w, playlists, http.StatusOK)
}

func (h *PlaylistHandler) GetPlaylistTracks(w http.ResponseWriter, r *http.Request) {
	playlistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	limit, cursorPos := parsePaginationPos(r)
	tracks, err := h.db.GetPlaylistTracks(r.Context(), sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        limit,
		Column3:      cursorPos,
	})
	if err != nil {
		h.logger.Error("failed to get playlist tracks", zap.Error(err))
		h.returns.ReturnError(w, "failed to get playlist tracks", http.StatusInternalServerError)
		return
	}

	for i := range tracks {
		tracks[i].PathInFileStorage = convertPathToFileURL(tracks[i].PathInFileStorage)
		applyDefaultImageIfEmpty(&tracks[i].ImagePath, "music")
	}

	h.returns.ReturnJSON(w, tracks, http.StatusOK)
}

func (h *PlaylistHandler) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns, h.logger) {
		return
	}

	// Get required fields
	rawOriginalName := r.FormValue("original_name")

	originalName, err := validateStringField(rawOriginalName, "original_name", 1, 200)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Optional description
	var description *string
	if descVal := r.FormValue("description"); descVal != "" {
		description = &descVal
	}

	// Optional is_public
	var isPublic *bool
	if publicVal := r.FormValue("is_public"); publicVal != "" {
		if publicBool, err := strconv.ParseBool(publicVal); err == nil {
			isPublic = &publicBool
		}
	}

	// Generate playlist ID for image upload
	playlistID := uuid.New().String()

	// Optional image upload
	imagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		consts.PicturesPlaylistFolder, playlistID, "image", h.logger, h.returns)
	if !ok {
		return
	}

	descText := optionalStringToPgtype(description)

	var publicBool pgtype.Bool
	if isPublic != nil {
		publicBool = pgtype.Bool{Bool: *isPublic, Valid: true}
	}

	if err := h.db.CreatePlaylist(r.Context(), sqlhandler.CreatePlaylistParams{
		FromUser:     userUUID,
		OriginalName: originalName,
		Description:  descText,
		IsPublic:     publicBool,
		ImagePath:    imagePath,
	}); err != nil {
		h.logger.Error("failed to create playlist", zap.Error(err))
		// If playlist creation fails and image was uploaded, try to clean up
		if imagePath.Valid {
			cleanupImage(r.Context(), h.fileStorage, consts.PicturesPlaylistFolder, playlistID, h.logger)
		}
		h.returns.ReturnError(w, "failed to create playlist", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "playlist created", http.StatusCreated)
}

type updatePlaylistRequest struct {
	OriginalName string  `json:"original_name" validate:"required,max=255"`
	Description  *string `json:"description"`
	IsPublic     *bool   `json:"is_public"`
}

func (h *PlaylistHandler) UpdatePlaylist(w http.ResponseWriter, r *http.Request) {
	userUUID, playlistUUID, ok := h.checkPlaylistOwnership(w, r)
	if !ok {
		return
	}

	body, ok := decodeBody[updatePlaylistRequest](w, r, h.returns)
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

	var isPublic pgtype.Bool
	if body.IsPublic != nil {
		isPublic = pgtype.Bool{Bool: *body.IsPublic, Valid: true}
	}

	if err := h.db.UpdatePlaylist(r.Context(), sqlhandler.UpdatePlaylistParams{
		UserUuid:     userUUID,
		Uuid:         playlistUUID,
		OriginalName: originalName,
		Description:  description,
		IsPublic:     isPublic,
	}); err != nil {
		h.logger.Error("failed to update playlist", zap.Error(err))
		h.returns.ReturnError(w, "failed to update playlist", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "playlist updated", http.StatusOK)
}

func (h *PlaylistHandler) DeletePlaylist(w http.ResponseWriter, r *http.Request) {
	userUUID, playlistUUID, ok := h.checkPlaylistOwnership(w, r)
	if !ok {
		return
	}

	if err := h.db.DeletePlaylist(r.Context(), sqlhandler.DeletePlaylistParams{UserUuid: userUUID, Uuid: playlistUUID}); err != nil {
		h.logger.Error("failed to delete playlist", zap.Error(err))
		h.returns.ReturnError(w, "failed to delete playlist", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "playlist deleted", http.StatusOK)
}

type addTrackRequest struct {
	MusicUUID string `json:"music_uuid" validate:"required"`
}

func (h *PlaylistHandler) AddTrackToPlaylist(w http.ResponseWriter, r *http.Request) {
	_, playlistUUID, ok := h.checkPlaylistOwnership(w, r)
	if !ok {
		return
	}

	body, ok := decodeBody[addTrackRequest](w, r, h.returns)
	if !ok {
		h.returns.ReturnError(w, "invalid inputs", http.StatusBadRequest)
		return
	}

	musicUUID, err := uuidToPgtype(body.MusicUUID)
	if err != nil {
		h.returns.ReturnError(w, "invalid music_uuid", http.StatusBadRequest)
		return
	}

	if err := h.db.AddTrackToPlaylist(r.Context(), sqlhandler.AddTrackToPlaylistParams{
		MusicUuid:    musicUUID,
		PlaylistUuid: playlistUUID,
	}); err != nil {
		h.logger.Error("failed to add track to playlist", zap.Error(err))
		h.returns.ReturnError(w, "failed to add track to playlist", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "track added to playlist", http.StatusCreated)
}

func (h *PlaylistHandler) RemoveTrackFromPlaylist(w http.ResponseWriter, r *http.Request) {
	userUUID, playlistUUID, ok := h.checkPlaylistOwnership(w, r)
	if !ok {
		return
	}

	// musicUuid path param refers to the music UUID (matches query's music_uuid param)
	musicUUID, ok := parseUUID(r, "musicUuid")
	if !ok {
		h.returns.ReturnError(w, "invalid music uuid", http.StatusBadRequest)
		return
	}

	if err := h.db.RemoveTrackFromPlaylist(r.Context(), sqlhandler.RemoveTrackFromPlaylistParams{
		UserUuid:     userUUID,
		MusicUuid:    musicUUID,
		PlaylistUuid: playlistUUID,
	}); err != nil {
		h.logger.Error("failed to remove track from playlist", zap.Error(err))
		h.returns.ReturnError(w, "failed to remove track from playlist", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "track removed from playlist", http.StatusOK)
}

func (h *PlaylistHandler) UpdatePlaylistImage(w http.ResponseWriter, r *http.Request) {
	userUUID, playlistUUID, ok := h.checkPlaylistOwnership(w, r)
	if !ok {
		return
	}

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns, h.logger) {
		return
	}

	imageID := uuid.UUID(playlistUUID.Bytes).String()

	imagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		consts.PicturesPlaylistFolder, imageID, "image", h.logger, h.returns)
	if !ok {
		return
	}

	if !imagePath.Valid {
		h.returns.ReturnError(w, "image file required", http.StatusBadRequest)
		return
	}

	// Update
	if err := h.db.UpdatePlaylistImage(r.Context(), sqlhandler.UpdatePlaylistImageParams{
		UserUuid:  userUUID,
		Uuid:      playlistUUID,
		ImagePath: imagePath,
	}); err != nil {
		h.logger.Error("failed to update playlist image", zap.Error(err))
		h.returns.ReturnError(w, "failed to update playlist image", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "playlist image updated", http.StatusOK)
}

type updateTrackPositionRequest struct {
	Position int32 `json:"position" validate:"gte=0"`
}

func (h *PlaylistHandler) UpdateTrackPosition(w http.ResponseWriter, r *http.Request) {
	userUUID, playlistUUID, ok := h.checkPlaylistOwnership(w, r)
	if !ok {
		return
	}

	// trackUuid is the playlist_track row UUID
	trackUUID, ok := parseUUID(r, "trackUuid")
	if !ok {
		h.returns.ReturnError(w, "invalid track uuid", http.StatusBadRequest)
		return
	}

	body, ok := decodeBody[updateTrackPositionRequest](w, r, h.returns)
	if !ok {
		return
	}

	if err := h.db.UpdateTrackPosition(r.Context(), sqlhandler.UpdateTrackPositionParams{
		UserUuid:     userUUID,
		PlaylistUuid: playlistUUID,
		Uuid:         trackUUID,
		Position:     body.Position,
	}); err != nil {
		h.logger.Error("failed to update track position", zap.Error(err))
		h.returns.ReturnError(w, "failed to update track position", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "track position updated", http.StatusOK)
}
