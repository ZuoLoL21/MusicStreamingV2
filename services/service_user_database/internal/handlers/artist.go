package handlers

import (
	"backend/internal/di"
	"backend/internal/storage"
	sqlhandler "backend/sql/sqlc"
	"net/http"

	libsdi "libs/di"
	libsmiddleware "libs/middleware"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type ArtistHandler struct {
	logger      *zap.Logger
	config      *di.Config
	returns     *libsdi.ReturnManager
	db          DB
	fileStorage storage.FileStorageClient
}

func NewArtistHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, db DB, fileStorage storage.FileStorageClient) *ArtistHandler {
	return &ArtistHandler{
		logger:      logger,
		config:      config,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
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
	logger := libsmiddleware.GetLogger(r.Context())

	limit, cursorName, cursorTS := parsePaginationAlpha(r)
	artists, err := h.db.GetArtistsAlphabetically(r.Context(), sqlhandler.GetArtistsAlphabeticallyParams{
		Limit:      limit,
		ArtistName: cursorName,
		Column3:    cursorTS,
	})
	if err != nil {
		logger.Warn("unable to retrieve artists",
			zap.Int32("limit", limit),
			zap.String("cursor_name", cursorName),
			zap.Error(err))
		h.returns.ReturnError(w, "unable to list artists", http.StatusInternalServerError)
		return
	}
	logger.Debug("retrieved artists alphabetically",
		zap.Int("count", len(artists)))
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

	applyDefaultImageIfEmpty(&artist.ProfileImagePath, h.fileStorage, "artist")
	h.returns.ReturnJSON(w, artist, http.StatusOK)
}

func (h *ArtistHandler) CreateArtist(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns) {
		return
	}

	artistName := r.FormValue("artist_name")
	if artistName == "" {
		h.returns.ReturnError(w, "artist_name required", http.StatusBadRequest)
		return
	}
	var bio *string
	if bioVal := r.FormValue("bio"); bioVal != "" {
		bio = &bioVal
	}

	// Generate artist ID
	artistID := uuid.New().String()

	profileImagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		"pictures-artist", artistID, "image", h.logger, h.returns)
	if !ok {
		return
	}

	if !profileImagePath.Valid {
		profileImagePath.String = h.fileStorage.GetDefaultArtistImageURL()
		profileImagePath.Valid = true
	}

	bioText := optionalStringToPgtype(bio)

	logger := libsmiddleware.GetLogger(r.Context())

	if err := h.db.CreateArtist(r.Context(), sqlhandler.CreateArtistParams{
		UserUuid:         userUUID,
		ArtistName:       artistName,
		Bio:              bioText,
		ProfileImagePath: profileImagePath,
	}); err != nil {
		logger.Error("database operation failed",
			zap.String("operation", "CreateArtist"),
			zap.Error(err),
			zap.String("artist_uuid", artistID),
			zap.String("artist_name", artistName))

		if profileImagePath.Valid {
			cleanupImage(r.Context(), h.fileStorage, "pictures-artist", artistID, h.logger)
		}
		h.returns.ReturnError(w, "unable to create artist profile", http.StatusInternalServerError)
		return
	}

	logger.Info("artist created successfully",
		zap.String("artist_uuid", artistID),
		zap.String("artist_name", artistName))
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

	bio := optionalStringToPgtype(body.Bio)

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

func (h *ArtistHandler) UpdateArtistPicture(w http.ResponseWriter, r *http.Request) {
	artistUUID, ok := h.checkArtistAccess(w, r, sqlhandler.ArtistMemberRoleManager)
	if !ok {
		return
	}

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns) {
		return
	}

	// Artist UUID
	imageID := uuid.UUID(artistUUID.Bytes).String()

	profileImagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		"pictures-artist", imageID, "image", h.logger, h.returns)
	if !ok {
		return
	}

	if !profileImagePath.Valid {
		h.returns.ReturnError(w, "image file required", http.StatusBadRequest)
		return
	}

	// Update
	if err := h.db.UpdateArtistPicture(r.Context(), sqlhandler.UpdateArtistPictureParams{
		Uuid:             artistUUID,
		ProfileImagePath: profileImagePath,
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
