package handlers

import (
	"net/http"
	"strconv"

	"backend/internal/di"
	"backend/internal/storage"

	sqlhandler "backend/sql/sqlc"
	libsdi "libs/di"
	libsmiddleware "libs/middleware"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type MusicHandler struct {
	logger      *zap.Logger
	config      *di.Config
	returns     *libsdi.ReturnManager
	db          DB
	fileStorage storage.FileStorageClient
}

func NewMusicHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, db DB, fileStorage storage.FileStorageClient) *MusicHandler {
	return &MusicHandler{
		logger:      logger,
		config:      config,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
	}
}

// checkMusicAccess parses the music UUID from the route, fetches the track to
// resolve its artist, and verifies the calling user has at least the given role.
func (h *MusicHandler) checkMusicAccess(w http.ResponseWriter, r *http.Request, role sqlhandler.ArtistMemberRole) (musicUUID pgtype.UUID, ok bool) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	musicUUID, ok = parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	music, err := h.db.GetMusic(r.Context(), musicUUID)
	if err != nil {
		h.returns.ReturnError(w, "music not found", http.StatusNotFound)
		ok = false
		return
	}

	if !checkArtistRole(r.Context(), h.db, music.FromArtist, userUUID, role) {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		ok = false
	}

	return
}

func (h *MusicHandler) GetMusic(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	musicUUID, ok := parseUUID(r, "uuid")
	if !ok {
		logger.Warn("invalid music uuid in request")
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	music, err := h.db.GetMusic(r.Context(), musicUUID)
	if err != nil {
		logger.Warn("music not found",
			zap.String("music_uuid", uuidToString(musicUUID)),
			zap.Error(err))
		h.returns.ReturnError(w, "unable to retrieve music", http.StatusNotFound)
		return
	}

	music.PathInFileStorage = h.fileStorage.BuildPublicURL(music.PathInFileStorage)
	applyDefaultImageIfEmpty(&music.ImagePath, h.fileStorage, "music")

	logger.Debug("music retrieved successfully",
		zap.String("music_uuid", uuidToString(musicUUID)))
	h.returns.ReturnJSON(w, music, http.StatusOK)
}

func (h *MusicHandler) GetMusicForArtist(w http.ResponseWriter, r *http.Request) {
	artistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	logger := libsmiddleware.GetLogger(r.Context())

	limit, cursorTS, cursorID := parsePagination(r)
	music, err := h.db.GetMusicForArtist(r.Context(), sqlhandler.GetMusicForArtistParams{
		FromArtist: artistUUID,
		Limit:      limit,
		Column3:    cursorTS,
		Uuid:       cursorID,
	})
	if err != nil {
		logger.Warn("unable to retrieve music for artist",
			zap.String("artist_uuid", uuidToString(artistUUID)),
			zap.Int32("limit", limit),
			zap.Error(err))
		h.returns.ReturnError(w, "unable to retrieve music catalog", http.StatusInternalServerError)
		return
	}

	for i := range music {
		music[i].PathInFileStorage = h.fileStorage.BuildPublicURL(music[i].PathInFileStorage)
		applyDefaultImageIfEmpty(&music[i].ImagePath, h.fileStorage, "music")
	}

	logger.Debug("retrieved music for artist",
		zap.String("artist_uuid", uuidToString(artistUUID)),
		zap.Int("count", len(music)))
	h.returns.ReturnJSON(w, music, http.StatusOK)
}

func (h *MusicHandler) GetMusicForAlbum(w http.ResponseWriter, r *http.Request) {
	albumUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	limit, cursorTS, cursorID := parsePagination(r)
	music, err := h.db.GetMusicForAlbum(r.Context(), sqlhandler.GetMusicForAlbumParams{
		InAlbum: albumUUID,
		Limit:   limit,
		Column3: cursorTS,
		Uuid:    cursorID,
	})
	if err != nil {
		h.logger.Error("failed to get music for album", zap.Error(err))
		h.returns.ReturnError(w, "failed to get music", http.StatusInternalServerError)
		return
	}

	for i := range music {
		music[i].PathInFileStorage = h.fileStorage.BuildPublicURL(music[i].PathInFileStorage)
		applyDefaultImageIfEmpty(&music[i].ImagePath, h.fileStorage, "music")
	}

	h.returns.ReturnJSON(w, music, http.StatusOK)
}

func (h *MusicHandler) GetMusicForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	logger := libsmiddleware.GetLogger(r.Context())

	limit, cursorTS, cursorID := parsePagination(r)
	music, err := h.db.GetMusicForUser(r.Context(), sqlhandler.GetMusicForUserParams{
		UploadedBy: userUUID,
		Limit:      limit,
		Column3:    cursorTS,
		Uuid:       cursorID,
	})
	if err != nil {
		logger.Warn("unable to retrieve music for user",
			zap.String("uploader_uuid", uuidToString(userUUID)),
			zap.Int32("limit", limit),
			zap.Error(err))
		h.returns.ReturnError(w, "unable to retrieve music catalog", http.StatusInternalServerError)
		return
	}

	for i := range music {
		music[i].PathInFileStorage = h.fileStorage.BuildPublicURL(music[i].PathInFileStorage)
		applyDefaultImageIfEmpty(&music[i].ImagePath, h.fileStorage, "music")
	}

	logger.Debug("retrieved music for user",
		zap.String("uploader_uuid", uuidToString(userUUID)),
		zap.Int("count", len(music)))
	h.returns.ReturnJSON(w, music, http.StatusOK)
}

func (h *MusicHandler) CreateMusic(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 100, h.returns, h.logger) {
		return
	}

	artistUUIDStr := r.FormValue("artist_uuid")
	if artistUUIDStr == "" {
		h.returns.ReturnError(w, "artist_uuid required", http.StatusBadRequest)
		return
	}

	songName := r.FormValue("song_name")
	if songName == "" {
		h.returns.ReturnError(w, "song_name required", http.StatusBadRequest)
		return
	}

	durationStr := r.FormValue("duration_seconds")
	if durationStr == "" {
		h.returns.ReturnError(w, "duration_seconds required", http.StatusBadRequest)
		return
	}

	durationSeconds, err := strconv.Atoi(durationStr)
	if err != nil || durationSeconds <= 0 {
		h.returns.ReturnError(w, "invalid duration_seconds", http.StatusBadRequest)
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

	var inAlbum pgtype.UUID
	inAlbumStr := r.FormValue("in_album")
	if inAlbumStr != "" {
		if err := inAlbum.Scan(inAlbumStr); err != nil {
			h.returns.ReturnError(w, "invalid in_album uuid", http.StatusBadRequest)
			return
		}
	}

	// Generate music ID
	musicID := uuid.New().String()

	// Update
	audioURL, ok := uploadAudioFromForm(r.Context(), w, r, h.fileStorage, musicID, "audio", h.logger, h.returns)
	if !ok {
		return
	}

	logger := libsmiddleware.GetLogger(r.Context())

	if err := h.db.CreateMusic(r.Context(), sqlhandler.CreateMusicParams{
		FromArtist:        artistUUID,
		UploadedBy:        userUUID,
		InAlbum:           inAlbum,
		SongName:          songName,
		PathInFileStorage: audioURL,
		DurationSeconds:   int32(durationSeconds),
	}); err != nil {
		logger.Error("database operation failed",
			zap.String("operation", "CreateMusic"),
			zap.Error(err),
			zap.String("music_uuid", musicID),
			zap.String("artist_uuid", uuidToString(artistUUID)),
			zap.String("song_name", songName))

		cleanupAudio(r.Context(), h.fileStorage, musicID, h.logger)

		h.returns.ReturnError(w, "unable to create music", http.StatusInternalServerError)
		return
	}

	logger.Info("music created successfully",
		zap.String("music_uuid", musicID),
		zap.String("artist_uuid", uuidToString(artistUUID)),
		zap.String("song_name", songName))
	h.returns.ReturnText(w, "music created", http.StatusCreated)
}

type updateMusicDetailsRequest struct {
	SongName string  `json:"song_name" validate:"required,max=255"`
	InAlbum  *string `json:"in_album"`
}

func (h *MusicHandler) UpdateMusicDetails(w http.ResponseWriter, r *http.Request) {
	musicUUID, ok := h.checkMusicAccess(w, r, sqlhandler.ArtistMemberRoleManager)
	if !ok {
		return
	}

	body, ok := decodeBody[updateMusicDetailsRequest](w, r, h.returns)
	if !ok {
		return
	}

	var inAlbum pgtype.UUID
	if body.InAlbum != nil {
		if err := inAlbum.Scan(*body.InAlbum); err != nil {
			h.returns.ReturnError(w, "invalid in_album uuid", http.StatusBadRequest)
			return
		}
	}

	if err := h.db.UpdateMusicDetails(r.Context(), sqlhandler.UpdateMusicDetailsParams{
		Uuid:     musicUUID,
		SongName: body.SongName,
		InAlbum:  inAlbum,
	}); err != nil {
		h.logger.Error("failed to update music details", zap.Error(err))
		h.returns.ReturnError(w, "failed to update music details", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "music details updated", http.StatusOK)
}

func (h *MusicHandler) UpdateMusicStorage(w http.ResponseWriter, r *http.Request) {
	musicUUID, ok := h.checkMusicAccess(w, r, sqlhandler.ArtistMemberRoleManager)
	if !ok {
		return
	}

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 100, h.returns, h.logger) {
		return
	}

	durationStr := r.FormValue("duration_seconds")
	if durationStr == "" {
		h.returns.ReturnError(w, "duration_seconds required", http.StatusBadRequest)
		return
	}

	durationSeconds, err := strconv.Atoi(durationStr)
	if err != nil || durationSeconds <= 0 {
		h.returns.ReturnError(w, "invalid duration_seconds", http.StatusBadRequest)
		return
	}

	// Use music UUID as deterministic audio ID (same filename on every update)
	musicID := uuid.UUID(musicUUID.Bytes).String()

	// Update
	audioURL, ok := uploadAudioFromForm(r.Context(), w, r, h.fileStorage, musicID, "audio", h.logger, h.returns)
	if !ok {
		return
	}

	if err := h.db.UpdateMusicStorage(r.Context(), sqlhandler.UpdateMusicStorageParams{
		Uuid:              musicUUID,
		PathInFileStorage: audioURL,
		DurationSeconds:   int32(durationSeconds),
	}); err != nil {
		h.logger.Error("failed to update music storage", zap.Error(err))
		h.returns.ReturnError(w, "failed to update music storage", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "music storage updated", http.StatusOK)
}

func (h *MusicHandler) DeleteMusic(w http.ResponseWriter, r *http.Request) {
	musicUUID, ok := h.checkMusicAccess(w, r, sqlhandler.ArtistMemberRoleOwner)
	if !ok {
		return
	}

	if err := h.db.DeleteMusic(r.Context(), musicUUID); err != nil {
		h.logger.Error("failed to delete music", zap.Error(err))
		h.returns.ReturnError(w, "failed to delete music", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "music deleted", http.StatusOK)
}

func (h *MusicHandler) IncrementPlayCount(w http.ResponseWriter, r *http.Request) {
	musicUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	if err := h.db.IncrementPlayCount(r.Context(), musicUUID); err != nil {
		h.logger.Error("failed to increment play count", zap.Error(err))
		h.returns.ReturnError(w, "failed to increment play count", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "play count incremented", http.StatusOK)
}

type addListeningHistoryRequest struct {
	ListenDurationSeconds *int32   `json:"listen_duration_seconds"`
	CompletionPercentage  *float64 `json:"completion_percentage"`
}

func (h *MusicHandler) AddListeningHistoryEntry(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	musicUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	body, ok := decodeBody[addListeningHistoryRequest](w, r, h.returns)
	if !ok {
		return
	}

	var duration pgtype.Int4
	if body.ListenDurationSeconds != nil {
		duration = pgtype.Int4{Int32: *body.ListenDurationSeconds, Valid: true}
	}

	var completion pgtype.Numeric
	if body.CompletionPercentage != nil {
		_ = completion.Scan(*body.CompletionPercentage)
	}

	if err := h.db.AddListeningHistoryEntry(r.Context(), sqlhandler.AddListeningHistoryEntryParams{
		UserUuid:              userUUID,
		MusicUuid:             musicUUID,
		ListenDurationSeconds: duration,
		CompletionPercentage:  completion,
	}); err != nil {
		h.logger.Error("failed to add listening history", zap.Error(err))
		h.returns.ReturnError(w, "failed to add listening history", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "listening history recorded", http.StatusCreated)
}
