package handlers

import (
	"backend/internal/consts"
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
	config      *di.Config
	returns     *libsdi.ReturnManager
	db          consts.DB
	fileStorage storage.FileStorageClient
}

func NewArtistHandler(config *di.Config, returns *libsdi.ReturnManager, db consts.DB, fileStorage storage.FileStorageClient) *ArtistHandler {
	return &ArtistHandler{
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

	for i := range artists {
		applyDefaultImageIfEmpty(&artists[i].ProfileImagePath, "artist")
	}

	logger.Debug("retrieved artists alphabetically",
		zap.Int("count", len(artists)))
	h.returns.ReturnJSON(w, artists, http.StatusOK)
}

func (h *ArtistHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	artistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	artist, err := h.db.GetArtist(r.Context(), artistUUID)
	if handleDBError(w, err, "artist not found", logger, h.returns) {
		return
	}

	applyDefaultImageIfEmpty(&artist.ProfileImagePath, "artist")
	h.returns.ReturnJSON(w, artist, http.StatusOK)
}

func (h *ArtistHandler) CreateArtist(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns) {
		return
	}

	rawArtistName := r.FormValue("artist_name")

	artistName, err := validateStringField(rawArtistName, "artist_name", 1, 200)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var bio *string
	if bioVal := r.FormValue("bio"); bioVal != "" {
		bio = &bioVal
	}

	// Generate artist ID
	artistID := uuid.New().String()

	profileImagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		consts.PicturesArtistFolder, artistID, "image", h.returns)
	if !ok {
		return
	}

	bioText := optionalStringToPgtype(bio)

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
			cleanupImage(r.Context(), h.fileStorage, consts.PicturesArtistFolder, artistID)
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
	logger := libsmiddleware.GetLogger(r.Context())

	artistUUID, ok := h.checkArtistAccess(w, r, sqlhandler.ArtistMemberRoleManager)
	if !ok {
		return
	}

	body, ok := decodeBody[updateArtistProfileRequest](w, r, h.returns)
	if !ok {
		return
	}

	artistName, err := validateStringField(body.ArtistName, "artist_name", 1, 200)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	bio := optionalStringToPgtype(body.Bio)

	if err := h.db.UpdateArtistProfile(r.Context(), sqlhandler.UpdateArtistProfileParams{
		Uuid:       artistUUID,
		ArtistName: artistName,
		Bio:        bio,
	}); err != nil {
		logger.Error("failed to update artist profile", zap.Error(err))
		h.returns.ReturnError(w, "failed to update artist profile", http.StatusInternalServerError)
		return
	}

	logger.Info("artist profile updated successfully",
		zap.String("artist_uuid", uuidToString(artistUUID)))
	h.returns.ReturnText(w, "artist profile updated", http.StatusOK)
}

func (h *ArtistHandler) UpdateArtistPicture(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	artistUUID, ok := h.checkArtistAccess(w, r, sqlhandler.ArtistMemberRoleManager)
	if !ok {
		return
	}

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns) {
		return
	}

	imageID := uuid.UUID(artistUUID.Bytes).String()

	// Upload
	profileImagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		consts.PicturesArtistFolder, imageID, "image", h.returns)
	if !ok {
		return
	}

	if !profileImagePath.Valid {
		h.returns.ReturnError(w, "image file required", http.StatusBadRequest)
		return
	}

	// Update database
	if err := h.db.UpdateArtistPicture(r.Context(), sqlhandler.UpdateArtistPictureParams{
		Uuid:             artistUUID,
		ProfileImagePath: profileImagePath,
	}); err != nil {
		logger.Error("failed to update artist picture in database", zap.Error(err))
		h.returns.ReturnError(w, "failed to update artist picture", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "artist picture updated", http.StatusOK)
}

func (h *ArtistHandler) GetUsersRepresentingArtist(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	artistUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	members, err := h.db.GetUsersRepresentingArtist(r.Context(), artistUUID)
	if err != nil {
		logger.Warn("failed to get artist members",
			zap.String("artist_uuid", uuidToString(artistUUID)),
			zap.Error(err))
		h.returns.ReturnError(w, "failed to get artist members", http.StatusInternalServerError)
		return
	}

	logger.Debug("artist members retrieved successfully",
		zap.String("artist_uuid", uuidToString(artistUUID)),
		zap.Int("count", len(members)))
	h.returns.ReturnJSON(w, members, http.StatusOK)
}

type addUserToArtistRequest struct {
	Role sqlhandler.ArtistMemberRole `json:"role" validate:"required"`
}

func (h *ArtistHandler) AddUserToArtist(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	artistUUID, targetUserUUID, ok := h.checkArtistAccessWithTarget(w, r, sqlhandler.ArtistMemberRoleOwner)
	if !ok {
		return
	}

	body, ok := decodeBody[addUserToArtistRequest](w, r, h.returns)
	if !ok {
		return
	}

	members, err := h.db.GetUsersRepresentingArtist(r.Context(), artistUUID)
	if err != nil {
		logger.Error("failed to get artist members", zap.Error(err))
		h.returns.ReturnError(w, "failed to check artist membership", http.StatusInternalServerError)
		return
	}
	for _, member := range members {
		if member.Uuid.Bytes == targetUserUUID.Bytes {
			h.returns.ReturnError(w, "user is already a member of this artist", http.StatusConflict)
			return
		}
	}

	if err := h.db.AddUserToArtist(r.Context(), sqlhandler.AddUserToArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   targetUserUUID,
		Role:       body.Role,
	}); err != nil {
		logger.Error("failed to add user to artist", zap.Error(err))
		h.returns.ReturnError(w, "failed to add user to artist", http.StatusInternalServerError)
		return
	}

	logger.Info("user added to artist successfully",
		zap.String("artist_uuid", uuidToString(artistUUID)),
		zap.String("user_uuid", uuidToString(targetUserUUID)))
	h.returns.ReturnText(w, "user added to artist", http.StatusCreated)
}

func (h *ArtistHandler) RemoveUserFromArtist(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	artistUUID, targetUserUUID, ok := h.checkArtistAccessWithTarget(w, r, sqlhandler.ArtistMemberRoleOwner)
	if !ok {
		return
	}

	if err := h.db.RemoveUserFromArtist(r.Context(), sqlhandler.RemoveUserFromArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   targetUserUUID,
	}); err != nil {
		logger.Error("failed to remove user from artist", zap.Error(err))
		h.returns.ReturnError(w, "failed to remove user from artist", http.StatusInternalServerError)
		return
	}

	logger.Info("user removed from artist successfully",
		zap.String("artist_uuid", uuidToString(artistUUID)),
		zap.String("user_uuid", uuidToString(targetUserUUID)))
	h.returns.ReturnText(w, "user removed from artist", http.StatusOK)
}

type changeUserRoleRequest struct {
	Role sqlhandler.ArtistMemberRole `json:"role" validate:"required"`
}

func (h *ArtistHandler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

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
		logger.Error("failed to change user role", zap.Error(err))
		h.returns.ReturnError(w, "failed to change user role", http.StatusInternalServerError)
		return
	}

	logger.Info("user role changed successfully",
		zap.String("artist_uuid", uuidToString(artistUUID)),
		zap.String("user_uuid", uuidToString(targetUserUUID)),
		zap.String("new_role", string(body.Role)))
	h.returns.ReturnText(w, "role updated", http.StatusOK)
}
