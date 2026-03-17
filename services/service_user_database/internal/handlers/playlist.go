package handlers

import (
	"backend/internal/consts"
	"net/http"
	"strconv"
	"time"

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
	if handleDBError(w, err, "playlist not found", h.logger, h.returns) {
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
	if handleDBError(w, err, "playlist not found", h.logger, h.returns) {
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

	// Retry logic with exponential backoff to handle concurrent position conflicts
	var lastErr error
	backoffMs := consts.InitialRetryBackoffMs
	for attempt := 0; attempt < consts.MaxRetries; attempt++ {
		err := h.db.AddTrackToPlaylist(r.Context(), sqlhandler.AddTrackToPlaylistParams{
			MusicUuid:    musicUUID,
			PlaylistUuid: playlistUUID,
		})
		if err == nil {
			h.returns.ReturnText(w, "track added to playlist", http.StatusCreated)
			return
		}
		lastErr = err

		if attempt < consts.MaxRetries-1 {
			time.Sleep(time.Millisecond * time.Duration(backoffMs))
			backoffMs *= consts.RetryBackoffMultiplier
		}
	}

	h.logger.Error("failed to add track to playlist after retries",
		zap.Error(lastErr),
		zap.Int("attempts", consts.MaxRetries))
	h.returns.ReturnError(w, "failed to add track to playlist", http.StatusInternalServerError)
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

type reorderPlaylistTracksRequest struct {
	TrackOrder []string `json:"track_order" validate:"required"`
}

func (h *PlaylistHandler) ReorderPlaylistTracks(w http.ResponseWriter, r *http.Request) {
	userUUID, playlistUUID, ok := h.checkPlaylistOwnership(w, r)
	if !ok {
		return
	}

	body, ok := decodeBody[reorderPlaylistTracksRequest](w, r, h.returns)
	if !ok {
		return
	}

	if len(body.TrackOrder) == 0 {
		h.returns.ReturnError(w, "track_order cannot be empty", http.StatusBadRequest)
		return
	}

	musicUUIDs := make([]pgtype.UUID, len(body.TrackOrder))
	seenUUIDs := make(map[string]bool)

	for i, uuidStr := range body.TrackOrder {
		parsedUUID, err := uuidToPgtype(uuidStr)
		if err != nil {
			h.returns.ReturnError(w, "invalid track uuid in track_order", http.StatusBadRequest)
			return
		}

		if seenUUIDs[uuidStr] {
			h.returns.ReturnError(w, "duplicate track uuid in track_order", http.StatusBadRequest)
			return
		}
		seenUUIDs[uuidStr] = true

		musicUUIDs[i] = parsedUUID
	}

	existingTracks, err := h.db.GetPlaylistTracks(r.Context(), sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        10000,
		Column3:      -1,
	})
	if err != nil {
		h.logger.Error("failed to get playlist tracks for validation", zap.Error(err))
		h.returns.ReturnError(w, "failed to reorder playlist tracks", http.StatusInternalServerError)
		return
	}

	if len(existingTracks) != len(musicUUIDs) {
		h.returns.ReturnError(w, "track_order count must match playlist track count", http.StatusBadRequest)
		return
	}

	// Checks if all provided UUIDs exist in playlist
	existingUUIDs := make(map[string]bool)
	for _, track := range existingTracks {
		existingUUIDs[uuid.UUID(track.Uuid.Bytes).String()] = true
	}
	for uuidStr := range seenUUIDs {
		if !existingUUIDs[uuidStr] {
			h.returns.ReturnError(w, "track uuid not found in playlist", http.StatusBadRequest)
			return
		}
	}

	if err := h.db.ReorderPlaylistTracks(r.Context(), sqlhandler.ReorderPlaylistTracksParams{
		UserUuid:     userUUID,
		PlaylistUuid: playlistUUID,
		Column3:      musicUUIDs,
	}); err != nil {
		h.logger.Error("failed to reorder playlist tracks", zap.Error(err))
		h.returns.ReturnError(w, "failed to reorder playlist tracks", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "playlist tracks reordered", http.StatusOK)
}
