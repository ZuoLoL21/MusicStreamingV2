package handlers

import (
	"backend/internal/consts"
	"bytes"
	"encoding/json"
	"io"
	libsconsts "libs/consts"
	"net/http"
	"time"

	"backend/internal/di"
	"backend/internal/storage"

	sqlhandler "backend/sql/sqlc"
	libsdi "libs/di"
	libsmiddleware "libs/middleware"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type TagsHandler struct {
	logger      *zap.Logger
	config      *di.Config
	jwtHandler  *libsdi.JWTHandler
	returns     *libsdi.ReturnManager
	db          consts.DB
	fileStorage storage.FileStorageClient
	httpClient  *http.Client
}

func NewTagsHandler(logger *zap.Logger, config *di.Config, jwtHandler *libsdi.JWTHandler, returns *libsdi.ReturnManager, db consts.DB, fileStorage storage.FileStorageClient) *TagsHandler {
	return &TagsHandler{
		logger:      logger,
		config:      config,
		jwtHandler:  jwtHandler,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (h *TagsHandler) GetAllTags(w http.ResponseWriter, r *http.Request) {
	limit, cursorName := parsePaginationName(r)
	tags, err := h.db.GetAllTags(r.Context(), sqlhandler.GetAllTagsParams{
		Limit:   limit,
		Column2: cursorName,
	})
	if err != nil {
		h.logger.Error("failed to get tags", zap.Error(err))
		h.returns.ReturnError(w, "failed to get tags", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, tags, http.StatusOK)
}

func (h *TagsHandler) GetTag(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	name := mux.Vars(r)["name"]

	tag, err := h.db.GetTag(r.Context(), name)
	if handleDBError(w, err, "tag not found", logger, h.returns) {
		return
	}

	h.returns.ReturnJSON(w, tag, http.StatusOK)
}

func (h *TagsHandler) GetMusicForTag(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	limit, cursorTS, cursorID := parsePagination(r)
	music, err := h.db.GetMusicForTag(r.Context(), sqlhandler.GetMusicForTagParams{
		TagName: name,
		Limit:   limit,
		Column3: cursorTS,
		Uuid:    cursorID,
	})
	if err != nil {
		h.logger.Error("failed to get music for tag", zap.Error(err))
		h.returns.ReturnError(w, "failed to get music for tag", http.StatusInternalServerError)
		return
	}

	for i := range music {
		music[i].PathInFileStorage = convertPathToFileURL(music[i].PathInFileStorage)
		applyDefaultImageIfEmpty(&music[i].ImagePath, "music")
	}

	h.returns.ReturnJSON(w, music, http.StatusOK)
}

func (h *TagsHandler) GetTagsForMusic(w http.ResponseWriter, r *http.Request) {
	musicUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	limit, cursorName := parsePaginationName(r)
	tags, err := h.db.GetTagsForMusic(r.Context(), sqlhandler.GetTagsForMusicParams{
		MusicUuid: musicUUID,
		Limit:     limit,
		Column3:   cursorName,
	})
	if err != nil {
		h.logger.Error("failed to get tags for music", zap.Error(err))
		h.returns.ReturnError(w, "failed to get tags for music", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, tags, http.StatusOK)
}

type createTagRequest struct {
	TagName        string  `json:"tag_name" validate:"required,max=100"`
	TagDescription *string `json:"tag_description"`
}

func (h *TagsHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeBody[createTagRequest](w, r, h.returns)
	if !ok {
		return
	}

	var description pgtype.Text
	if body.TagDescription != nil {
		description = pgtype.Text{String: *body.TagDescription, Valid: true}
	}

	if err := h.db.CreateTag(r.Context(), sqlhandler.CreateTagParams{
		TagName:        body.TagName,
		TagDescription: description,
	}); err != nil {
		h.logger.Error("failed to create tag", zap.Error(err))
		h.returns.ReturnError(w, "failed to create tag", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "tag created", http.StatusCreated)
}

func (h *TagsHandler) checkTagMusicAccess(w http.ResponseWriter, r *http.Request) (musicUUID pgtype.UUID, ok bool) {
	logger := libsmiddleware.GetLogger(r.Context())

	musicUUID, ok = parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	music, err := h.db.GetMusic(r.Context(), musicUUID)
	if handleDBError(w, err, "music not found", logger, h.returns) {
		ok = false
		return
	}

	if !checkArtistRole(r.Context(), h.db, music.FromArtist, userUUID, sqlhandler.ArtistMemberRoleMember) {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		ok = false
	}

	return
}

func (h *TagsHandler) AssignTagToMusic(w http.ResponseWriter, r *http.Request) {
	musicUUID, ok := h.checkTagMusicAccess(w, r)
	if !ok {
		return
	}

	name := mux.Vars(r)["name"]

	if err := h.db.AssignTagToMusic(r.Context(), sqlhandler.AssignTagToMusicParams{
		MusicUuid: musicUUID,
		TagName:   name,
	}); err != nil {
		h.logger.Error("failed to assign tag to music", zap.Error(err))
		h.returns.ReturnError(w, "failed to assign tag to music", http.StatusInternalServerError)
		return
	}

	go h.syncThemeToClickHouse(musicUUID, name)

	h.returns.ReturnText(w, "tag assigned to music", http.StatusOK)
}

func (h *TagsHandler) RemoveTagFromMusic(w http.ResponseWriter, r *http.Request) {
	musicUUID, ok := h.checkTagMusicAccess(w, r)
	if !ok {
		return
	}

	name := mux.Vars(r)["name"]

	if err := h.db.RemoveTagFromMusic(r.Context(), sqlhandler.RemoveTagFromMusicParams{
		MusicUuid: musicUUID,
		TagName:   name,
	}); err != nil {
		h.logger.Error("failed to remove tag from music", zap.Error(err))
		h.returns.ReturnError(w, "failed to remove tag from music", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "tag removed from music", http.StatusOK)
}

// syncThemeToClickHouse sends theme data to the event ingestion service
func (h *TagsHandler) syncThemeToClickHouse(musicUUID pgtype.UUID, theme string) {
	if h.config.EventIngestionServiceURL == "" {
		return
	}

	musicUUIDStr := uuid.UUID(musicUUID.Bytes).String()

	payload := map[string]interface{}{
		"music_uuid": musicUUIDStr,
		"theme":      theme,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("failed to marshal theme payload", zap.Error(err))
		return
	}

	req, err := http.NewRequest("POST", h.config.EventIngestionServiceURL+"/events/theme", bytes.NewBuffer(jsonData))
	if err != nil {
		h.logger.Error("failed to create theme request", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json")

	serviceJWT, err := h.jwtHandler.GenerateJwt(
		libsconsts.JWTSubjectService,
		"system",
		2*time.Minute,
	)
	if err != nil {
		h.logger.Error("failed to generate service JWT", zap.Error(err))
		return
	}
	req.Header.Set("Authorization", "Bearer "+serviceJWT)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Warn("failed to sync theme to ClickHouse", zap.Error(err))
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		h.logger.Warn("theme sync failed", zap.Int("status", resp.StatusCode))
	} else {
		h.logger.Debug("theme synced to ClickHouse",
			zap.String("music_uuid", musicUUIDStr),
			zap.String("theme", theme))
	}
}
