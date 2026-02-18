package handlers

import (
	"net/http"

	"backend/internal/di"
	sqlhandler "backend/sql/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type MusicHandler struct {
	logger  *zap.Logger
	config  *di.Config
	returns *di.ReturnManager
	db      *sqlhandler.Queries
}

func NewMusicHandler(logger *zap.Logger, config *di.Config, returns *di.ReturnManager, db *sqlhandler.Queries) *MusicHandler {
	return &MusicHandler{
		logger:  logger,
		config:  config,
		returns: returns,
		db:      db,
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
	musicUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	music, err := h.db.GetMusic(r.Context(), musicUUID)
	if err != nil {
		h.returns.ReturnError(w, "music not found", http.StatusNotFound)
		return
	}

	h.returns.ReturnJSON(w, music, http.StatusOK)
}

func (h *MusicHandler) GetMusicForArtist(w http.ResponseWriter, r *http.Request) {
	artistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	music, err := h.db.GetMusicForArtist(r.Context(), artistUUID)
	if err != nil {
		h.logger.Error("failed to get music for artist", zap.Error(err))
		h.returns.ReturnError(w, "failed to get music", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, music, http.StatusOK)
}

func (h *MusicHandler) GetMusicForAlbum(w http.ResponseWriter, r *http.Request) {
	albumUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	music, err := h.db.GetMusicForAlbum(r.Context(), albumUUID)
	if err != nil {
		h.logger.Error("failed to get music for album", zap.Error(err))
		h.returns.ReturnError(w, "failed to get music", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, music, http.StatusOK)
}

func (h *MusicHandler) GetMusicForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	music, err := h.db.GetMusicForUser(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("failed to get music for user", zap.Error(err))
		h.returns.ReturnError(w, "failed to get music", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, music, http.StatusOK)
}

type createMusicRequest struct {
	ArtistUUID        string  `json:"artist_uuid" validate:"required"`
	InAlbum           *string `json:"in_album"`
	SongName          string  `json:"song_name" validate:"required,max=255"`
	PathInFileStorage string  `json:"path_in_file_storage" validate:"required"`
	DurationSeconds   int32   `json:"duration_seconds" validate:"gt=0"`
}

func (h *MusicHandler) CreateMusic(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	body, ok := decodeBody[createMusicRequest](w, r, h.returns)
	if !ok {
		return
	}

	artistUUID, err := uuidToPgtype(body.ArtistUUID)
	if err != nil {
		h.returns.ReturnError(w, "invalid artist_uuid", http.StatusBadRequest)
		return
	}

	if !checkArtistRole(r.Context(), h.db, artistUUID, userUUID, sqlhandler.ArtistMemberRoleMember) {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		return
	}

	var inAlbum pgtype.UUID
	if body.InAlbum != nil {
		if err := inAlbum.Scan(*body.InAlbum); err != nil {
			h.returns.ReturnError(w, "invalid in_album uuid", http.StatusBadRequest)
			return
		}
	}

	if err := h.db.CreateMusic(r.Context(), sqlhandler.CreateMusicParams{
		FromArtist:        artistUUID,
		UploadedBy:        userUUID,
		InAlbum:           inAlbum,
		SongName:          body.SongName,
		PathInFileStorage: body.PathInFileStorage,
		DurationSeconds:   body.DurationSeconds,
	}); err != nil {
		h.logger.Error("failed to create music", zap.Error(err))
		h.returns.ReturnError(w, "failed to create music", http.StatusInternalServerError)
		return
	}

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

type updateMusicStorageRequest struct {
	PathInFileStorage string `json:"path_in_file_storage" validate:"required"`
	DurationSeconds   int32  `json:"duration_seconds" validate:"gt=0"`
}

func (h *MusicHandler) UpdateMusicStorage(w http.ResponseWriter, r *http.Request) {
	musicUUID, ok := h.checkMusicAccess(w, r, sqlhandler.ArtistMemberRoleManager)
	if !ok {
		return
	}

	body, ok := decodeBody[updateMusicStorageRequest](w, r, h.returns)
	if !ok {
		return
	}

	if err := h.db.UpdateMusicStorage(r.Context(), sqlhandler.UpdateMusicStorageParams{
		Uuid:              musicUUID,
		PathInFileStorage: body.PathInFileStorage,
		DurationSeconds:   body.DurationSeconds,
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
