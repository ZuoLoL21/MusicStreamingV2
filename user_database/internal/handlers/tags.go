package handlers

import (
	"net/http"

	"backend/internal/di"
	sqlhandler "backend/sql/sqlc"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type TagsHandler struct {
	logger  *zap.Logger
	config  *di.Config
	returns *di.ReturnManager
	db      *sqlhandler.Queries
}

func NewTagsHandler(logger *zap.Logger, config *di.Config, returns *di.ReturnManager, db *sqlhandler.Queries) *TagsHandler {
	return &TagsHandler{
		logger:  logger,
		config:  config,
		returns: returns,
		db:      db,
	}
}

func (h *TagsHandler) GetAllTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.db.GetAllTags(r.Context())
	if err != nil {
		h.logger.Error("failed to get tags", zap.Error(err))
		h.returns.ReturnError(w, "failed to get tags", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, tags, http.StatusOK)
}

func (h *TagsHandler) GetTag(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	tag, err := h.db.GetTag(r.Context(), name)
	if err != nil {
		h.returns.ReturnError(w, "tag not found", http.StatusNotFound)
		return
	}

	h.returns.ReturnJSON(w, tag, http.StatusOK)
}

func (h *TagsHandler) GetMusicForTag(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	music, err := h.db.GetMusicForTag(r.Context(), name)
	if err != nil {
		h.logger.Error("failed to get music for tag", zap.Error(err))
		h.returns.ReturnError(w, "failed to get music for tag", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, music, http.StatusOK)
}

func (h *TagsHandler) GetTagsForMusic(w http.ResponseWriter, r *http.Request) {
	musicUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	tags, err := h.db.GetTagsForMusic(r.Context(), musicUUID)
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

func (h *TagsHandler) AssignTagToMusic(w http.ResponseWriter, r *http.Request) {
	musicUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
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

	h.returns.ReturnText(w, "tag assigned to music", http.StatusOK)
}

func (h *TagsHandler) RemoveTagFromMusic(w http.ResponseWriter, r *http.Request) {
	musicUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
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
