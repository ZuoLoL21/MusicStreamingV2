package handlers

import (
	"backend/internal/client"
	"backend/internal/consts"
	"backend/internal/di"
	"backend/internal/storage"
	sqlhandler "backend/sql/sqlc"
	"net/http"
	"time"

	libsdi "libs/di"
	libshelpers "libs/helpers"
	libsmiddleware "libs/middleware"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserHandler struct {
	logger         *zap.Logger
	config         *di.Config
	jwtHandler     *libsdi.JWTHandler
	returns        *libsdi.ReturnManager
	db             consts.DB
	fileStorage    storage.FileStorageClient
	clickhouseSync *client.ClickHouseSync
}

func NewUserHandler(logger *zap.Logger, config *di.Config, jwtHandler *libsdi.JWTHandler, returns *libsdi.ReturnManager, db consts.DB, fileStorage storage.FileStorageClient, clickhouseSync *client.ClickHouseSync) *UserHandler {
	return &UserHandler{
		logger:         logger,
		config:         config,
		jwtHandler:     jwtHandler,
		returns:        returns,
		db:             db,
		fileStorage:    fileStorage,
		clickhouseSync: clickhouseSync,
	}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	user, err := h.db.GetPublicUser(r.Context(), userUUID)
	if handleDBError(w, err, "user not found", logger, h.returns) {
		return
	}

	applyDefaultImageIfEmpty(&user.ProfileImagePath, "user")
	h.returns.ReturnJSON(w, user, http.StatusOK)
}

func (h *UserHandler) GetPublicUser(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	userUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	user, err := h.db.GetPublicUser(r.Context(), userUUID)
	if handleDBError(w, err, "user not found", logger, h.returns) {
		return
	}

	applyDefaultImageIfEmpty(&user.ProfileImagePath, "user")
	h.returns.ReturnJSON(w, user, http.StatusOK)
}

type updateProfileRequest struct {
	Username string  `json:"username" validate:"required,min=5"`
	Bio      *string `json:"bio"`
	Country  string  `json:"country" validate:"required,len=2"`
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	body, ok := decodeBody[updateProfileRequest](w, r, h.returns)
	if !ok {
		return
	}

	username, err := validateStringField(body.Username, "username", 5, 100)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	bio := optionalStringToPgtype(body.Bio)

	if err := h.db.UpdateProfile(r.Context(), sqlhandler.UpdateProfileParams{
		Uuid:     userUUID,
		Username: username,
		Bio:      bio,
		Country:  body.Country,
	}); err != nil {
		h.logger.Error("failed to update profile", zap.Error(err))
		h.returns.ReturnError(w, "failed to update profile", http.StatusInternalServerError)
		return
	}

	h.clickhouseSync.SyncUserDim(userUUID, body.Country, time.Now())

	h.returns.ReturnText(w, "profile updated", http.StatusOK)
}

type updateEmailRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
}

func (h *UserHandler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	body, ok := decodeBody[updateEmailRequest](w, r, h.returns)
	if !ok {
		return
	}

	hashedPassword, err := h.db.GetHashPassword(r.Context(), userUUID)
	if handleDBError(w, err, "user not found", logger, h.returns) {
		return
	}

	if !libshelpers.Verify(body.OldPassword, hashedPassword) {
		h.returns.ReturnError(w, "invalid password", http.StatusUnauthorized)
		return
	}

	if err := h.db.UpdateEmail(r.Context(), sqlhandler.UpdateEmailParams{
		Uuid:  userUUID,
		Email: body.Email,
	}); err != nil {
		h.logger.Error("failed to update email", zap.Error(err))
		h.returns.ReturnError(w, "failed to update email", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "email updated", http.StatusOK)
}

type updatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	body, ok := decodeBody[updatePasswordRequest](w, r, h.returns)
	if !ok {
		return
	}

	hashedPassword, err := h.db.GetHashPassword(r.Context(), userUUID)
	if handleDBError(w, err, "user not found", logger, h.returns) {
		return
	}

	if !libshelpers.Verify(body.OldPassword, hashedPassword) {
		h.returns.ReturnError(w, "invalid password", http.StatusUnauthorized)
		return
	}

	newHashed := libshelpers.Encode(body.NewPassword)
	if err := h.db.UpdatePassword(r.Context(), sqlhandler.UpdatePasswordParams{
		Uuid:           userUUID,
		HashedPassword: newHashed,
	}); err != nil {
		h.logger.Error("failed to update password", zap.Error(err))
		h.returns.ReturnError(w, "failed to update password", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "password updated", http.StatusOK)
}

func (h *UserHandler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns, h.logger) {
		return
	}

	imageID := uuid.UUID(userUUID.Bytes).String()

	profileImagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		consts.PicturesProfileFolder, imageID, "image", h.logger, h.returns)
	if !ok {
		return
	}

	if !profileImagePath.Valid {
		h.returns.ReturnError(w, "image file required", http.StatusBadRequest)
		return
	}

	// Update
	if err := h.db.UpdateImage(r.Context(), sqlhandler.UpdateImageParams{
		Uuid:             userUUID,
		ProfileImagePath: profileImagePath,
	}); err != nil {
		h.logger.Error("failed to update image", zap.Error(err))
		h.returns.ReturnError(w, "failed to update image", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "image updated", http.StatusOK)
}

func (h *UserHandler) GetArtistForUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	artists, err := h.db.GetArtistForUser(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("failed to get artists for user", zap.Error(err))
		h.returns.ReturnError(w, "failed to get artists", http.StatusInternalServerError)
		return
	}

	for i := range artists {
		applyDefaultImageIfEmpty(&artists[i].ProfileImagePath, "artist")
	}

	h.returns.ReturnJSON(w, artists, http.StatusOK)
}
