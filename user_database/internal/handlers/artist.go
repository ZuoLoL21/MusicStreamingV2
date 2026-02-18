package handlers

import (
	"net/http"

	"backend/internal/di"
	sqlhandler "backend/sql/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type ArtistHandler struct {
	logger  *zap.Logger
	config  *di.Config
	returns *di.ReturnManager
	db      DB
}

func NewArtistHandler(logger *zap.Logger, config *di.Config, returns *di.ReturnManager, db DB) *ArtistHandler {
	return &ArtistHandler{
		logger:  logger,
		config:  config,
		returns: returns,
		db:      db,
	}
}

// checkArtistAccess parses the artist UUID from the route and verifies the
// calling user has at least the given role on that artist.
// Used for just checking if a user can modify information about an artist
func (h *ArtistHandler) checkArtistAccess(w http.ResponseWriter, r *http.Request, role sqlhandler.ArtistMemberRole) (artistUUID pgtype.UUID, ok bool) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	artistUUID, ok = parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	if !checkArtistRole(r.Context(), h.db, artistUUID, userUUID, role) {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		ok = false
	}

	return
}

// checkArtistAccessWithTarget is like checkArtistAccess but also parses a
// second "userUuid" route parameter for operations that target another user.
// Used for retrieving an additional user to add/modify wrt an artist
func (h *ArtistHandler) checkArtistAccessWithTarget(w http.ResponseWriter, r *http.Request, role sqlhandler.ArtistMemberRole) (artistUUID pgtype.UUID, targetUserUUID pgtype.UUID, ok bool) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	artistUUID, ok = parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	targetUserUUID, ok = parseUUID(r, "userUuid")
	if !ok {
		h.returns.ReturnError(w, "invalid user uuid", http.StatusBadRequest)
		return
	}

	if !checkArtistRole(r.Context(), h.db, artistUUID, userUUID, role) {
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		ok = false
	}

	return
}

func (h *ArtistHandler) GetArtistsAlphabetically(w http.ResponseWriter, r *http.Request) {
	artists, err := h.db.GetArtistsAlphabetically(r.Context())
	if err != nil {
		h.logger.Error("failed to get artists", zap.Error(err))
		h.returns.ReturnError(w, "failed to get artists", http.StatusInternalServerError)
		return
	}
	h.returns.ReturnJSON(w, artists, http.StatusOK)
}

func (h *ArtistHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	artistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	artist, err := h.db.GetArtist(r.Context(), artistUUID)
	if err != nil {
		h.returns.ReturnError(w, "artist not found", http.StatusNotFound)
		return
	}

	h.returns.ReturnJSON(w, artist, http.StatusOK)
}

type createArtistRequest struct {
	ArtistName       string  `json:"artist_name" validate:"required,max=255"`
	Bio              *string `json:"bio"`
	ProfileImagePath *string `json:"profile_image_path"`
}

func (h *ArtistHandler) CreateArtist(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	body, ok := decodeBody[createArtistRequest](w, r, h.returns)
	if !ok {
		return
	}

	var bio pgtype.Text
	if body.Bio != nil {
		bio = pgtype.Text{String: *body.Bio, Valid: true}
	}

	var profileImagePath pgtype.Text
	if body.ProfileImagePath != nil {
		profileImagePath = pgtype.Text{String: *body.ProfileImagePath, Valid: true}
	}

	if err := h.db.CreateArtist(r.Context(), sqlhandler.CreateArtistParams{
		UserUuid:         userUUID,
		ArtistName:       body.ArtistName,
		Bio:              bio,
		ProfileImagePath: profileImagePath,
	}); err != nil {
		h.logger.Error("failed to create artist", zap.Error(err))
		h.returns.ReturnError(w, "failed to create artist", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "artist created", http.StatusCreated)
}

type updateArtistProfileRequest struct {
	ArtistName string  `json:"artist_name" validate:"required,max=255"`
	Bio        *string `json:"bio"`
}

func (h *ArtistHandler) UpdateArtistProfile(w http.ResponseWriter, r *http.Request) {
	artistUUID, ok := h.checkArtistAccess(w, r, sqlhandler.ArtistMemberRoleManager)
	if !ok {
		return
	}

	body, ok := decodeBody[updateArtistProfileRequest](w, r, h.returns)
	if !ok {
		return
	}

	var bio pgtype.Text
	if body.Bio != nil {
		bio = pgtype.Text{String: *body.Bio, Valid: true}
	}

	if err := h.db.UpdateArtistProfile(r.Context(), sqlhandler.UpdateArtistProfileParams{
		Uuid:       artistUUID,
		ArtistName: body.ArtistName,
		Bio:        bio,
	}); err != nil {
		h.logger.Error("failed to update artist profile", zap.Error(err))
		h.returns.ReturnError(w, "failed to update artist profile", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "artist profile updated", http.StatusOK)
}

type updateArtistPictureRequest struct {
	ProfileImagePath string `json:"profile_image_path" validate:"required"`
}

func (h *ArtistHandler) UpdateArtistPicture(w http.ResponseWriter, r *http.Request) {
	artistUUID, ok := h.checkArtistAccess(w, r, sqlhandler.ArtistMemberRoleManager)
	if !ok {
		return
	}

	body, ok := decodeBody[updateArtistPictureRequest](w, r, h.returns)
	if !ok {
		return
	}

	if err := h.db.UpdateArtistPicture(r.Context(), sqlhandler.UpdateArtistPictureParams{
		Uuid:             artistUUID,
		ProfileImagePath: pgtype.Text{String: body.ProfileImagePath, Valid: true},
	}); err != nil {
		h.logger.Error("failed to update artist picture", zap.Error(err))
		h.returns.ReturnError(w, "failed to update artist picture", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "artist picture updated", http.StatusOK)
}

func (h *ArtistHandler) GetUsersRepresentingArtist(w http.ResponseWriter, r *http.Request) {
	artistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	members, err := h.db.GetUsersRepresentingArtist(r.Context(), artistUUID)
	if err != nil {
		h.logger.Error("failed to get artist members", zap.Error(err))
		h.returns.ReturnError(w, "failed to get artist members", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, members, http.StatusOK)
}

type addUserToArtistRequest struct {
	Role sqlhandler.ArtistMemberRole `json:"role" validate:"required"`
}

func (h *ArtistHandler) AddUserToArtist(w http.ResponseWriter, r *http.Request) {
	artistUUID, targetUserUUID, ok := h.checkArtistAccessWithTarget(w, r, sqlhandler.ArtistMemberRoleOwner)
	if !ok {
		return
	}

	body, ok := decodeBody[addUserToArtistRequest](w, r, h.returns)
	if !ok {
		return
	}

	if err := h.db.AddUserToArtist(r.Context(), sqlhandler.AddUserToArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   targetUserUUID,
		Role:       body.Role,
	}); err != nil {
		h.logger.Error("failed to add user to artist", zap.Error(err))
		h.returns.ReturnError(w, "failed to add user to artist", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "user added to artist", http.StatusCreated)
}

func (h *ArtistHandler) RemoveUserFromArtist(w http.ResponseWriter, r *http.Request) {
	artistUUID, targetUserUUID, ok := h.checkArtistAccessWithTarget(w, r, sqlhandler.ArtistMemberRoleOwner)
	if !ok {
		return
	}

	if err := h.db.RemoveUserFromArtist(r.Context(), sqlhandler.RemoveUserFromArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   targetUserUUID,
	}); err != nil {
		h.logger.Error("failed to remove user from artist", zap.Error(err))
		h.returns.ReturnError(w, "failed to remove user from artist", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "user removed from artist", http.StatusOK)
}

type changeUserRoleRequest struct {
	Role sqlhandler.ArtistMemberRole `json:"role" validate:"required"`
}

func (h *ArtistHandler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	artistUUID, targetUserUUID, ok := h.checkArtistAccessWithTarget(w, r, sqlhandler.ArtistMemberRoleOwner)
	if !ok {
		return
	}

	body, ok := decodeBody[changeUserRoleRequest](w, r, h.returns)
	if !ok {
		return
	}

	if err := h.db.ChangeUserRole(r.Context(), sqlhandler.ChangeUserRoleParams{
		ArtistUuid: artistUUID,
		UserUuid:   targetUserUUID,
		Role:       body.Role,
	}); err != nil {
		h.logger.Error("failed to change user role", zap.Error(err))
		h.returns.ReturnError(w, "failed to change user role", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "role updated", http.StatusOK)
}
