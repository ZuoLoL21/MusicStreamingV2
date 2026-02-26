package handlers

import (
	"backend/internal/di"
	"backend/internal/storage"
	sqlhandler "backend/sql/sqlc"
	"fmt"
	"net/http"

	libsdi "libs/di"
	libshelpers "libs/helpers"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserHandler struct {
	logger      *zap.Logger
	config      *di.Config
	secrets     *libsdi.SecretsManager
	returns     *libsdi.ReturnManager
	db          DB
	fileStorage storage.FileStorageClient
}

func NewUserHandler(logger *zap.Logger, config *di.Config, secrets *libsdi.SecretsManager, returns *libsdi.ReturnManager, db DB, fileStorage storage.FileStorageClient) *UserHandler {
	return &UserHandler{
		logger:      logger,
		config:      config,
		secrets:     secrets,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
	}
}

type tokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *UserHandler) issueTokenPair(uuidStr string) tokenPair {
	_, priKey, kid := h.secrets.GetKeyInfo(h.config.JWTStorePath)
	access := libshelpers.GenerateJwt(libshelpers.JWTSubjectNormal, uuidStr, priKey, kid, h.config.JWTExpirationNormal)
	refresh := libshelpers.GenerateJwt(libshelpers.JWTSubjectRefresh, uuidStr, priKey, kid, h.config.JWTExpirationRefresh)
	return tokenPair{AccessToken: access, RefreshToken: refresh}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns) {
		return
	}

	username := r.FormValue("username")
	if username == "" || len(username) < 5 {
		h.returns.ReturnError(w, "username required (min 5 chars)", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	if email == "" {
		h.returns.ReturnError(w, "email required", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	if password == "" || len(password) < 8 {
		h.returns.ReturnError(w, "password required (min 8 chars)", http.StatusBadRequest)
		return
	}

	var bio *string
	if bioVal := r.FormValue("bio"); bioVal != "" {
		bio = &bioVal
	}

	hashedPassword := libshelpers.Encode(password)
	bioText := optionalStringToPgtype(bio)

	// Create user first without image to get the UUID
	createdUUID, err := h.db.CreateUser(r.Context(), sqlhandler.CreateUserParams{
		Username:       username,
		Email:          email,
		HashedPassword: hashedPassword,
		Bio:            bioText,
	})
	if err != nil {
		h.logger.Warn("failed to create user", zap.Error(err))
		h.returns.ReturnError(
			w,
			fmt.Sprintf("failed to create user, %v", err.Error()),
			http.StatusInternalServerError,
		)
		return
	}

	userID := uuid.UUID(createdUUID.Bytes).String()

	// Upload
	profileImagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		"pictures-profile", userID, "image", h.logger, h.returns)
	if !ok {
		return
	}

	if !profileImagePath.Valid {
		profileImagePath.String = h.fileStorage.GetDefaultProfileImageURL()
		profileImagePath.Valid = true
	}

	// Update
	if err := h.db.UpdateImage(r.Context(), sqlhandler.UpdateImageParams{
		Uuid:             createdUUID,
		ProfileImagePath: profileImagePath,
	}); err != nil {
		h.logger.Error("failed to update user image after creation", zap.Error(err))
		h.logger.Warn("user created but image update failed",
			zap.String("userID", userID),
			zap.Error(err))
	}

	uuidStr := uuid.UUID(createdUUID.Bytes).String()
	h.returns.ReturnJSON(w, h.issueTokenPair(uuidStr), http.StatusCreated)
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeBody[loginRequest](w, r, h.returns)
	if !ok {
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		h.returns.ReturnError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !libshelpers.Verify(body.Password, user.HashedPassword) {
		h.returns.ReturnError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	uuidStr := uuid.UUID(user.Uuid.Bytes).String()
	h.returns.ReturnJSON(w, h.issueTokenPair(uuidStr), http.StatusOK)
}

func (h *UserHandler) Renew(w http.ResponseWriter, r *http.Request) {
	uuidStr, ok := r.Context().Value(h.config.UserUUIDKey).(string)
	if !ok || uuidStr == "" {
		h.returns.ReturnError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	_, priKey, kid := h.secrets.GetKeyInfo(h.config.JWTStorePath)
	access := libshelpers.GenerateJwt(libshelpers.JWTSubjectNormal, uuidStr, priKey, kid, h.config.JWTExpirationNormal)
	h.returns.ReturnJSON(w, map[string]string{"access_token": access}, http.StatusOK)
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	user, err := h.db.GetPublicUser(r.Context(), userUUID)
	if err != nil {
		h.returns.ReturnError(w, "user not found", http.StatusNotFound)
		return
	}

	applyDefaultImageIfEmpty(&user.ProfileImagePath, h.fileStorage, "user")
	h.returns.ReturnJSON(w, user, http.StatusOK)
}

func (h *UserHandler) GetPublicUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseUUID(r, "uuid")
	if !ok {
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	user, err := h.db.GetPublicUser(r.Context(), userUUID)
	if err != nil {
		h.returns.ReturnError(w, "user not found", http.StatusNotFound)
		return
	}

	applyDefaultImageIfEmpty(&user.ProfileImagePath, h.fileStorage, "user")
	h.returns.ReturnJSON(w, user, http.StatusOK)
}

type updateProfileRequest struct {
	Username string  `json:"username" validate:"required,min=5"`
	Bio      *string `json:"bio"`
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

	bio := optionalStringToPgtype(body.Bio)

	if err := h.db.UpdateProfile(r.Context(), sqlhandler.UpdateProfileParams{
		Uuid:     userUUID,
		Username: body.Username,
		Bio:      bio,
	}); err != nil {
		h.logger.Error("failed to update profile", zap.Error(err))
		h.returns.ReturnError(w, "failed to update profile", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnText(w, "profile updated", http.StatusOK)
}

type updateEmailRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
}

func (h *UserHandler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	body, ok := decodeBody[updateEmailRequest](w, r, h.returns)
	if !ok {
		return
	}

	hashedPassword, err := h.db.GetHashPassword(r.Context(), userUUID)
	if err != nil {
		h.returns.ReturnError(w, "user not found", http.StatusNotFound)
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
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	body, ok := decodeBody[updatePasswordRequest](w, r, h.returns)
	if !ok {
		return
	}

	hashedPassword, err := h.db.GetHashPassword(r.Context(), userUUID)
	if err != nil {
		h.returns.ReturnError(w, "user not found", http.StatusNotFound)
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
	if !parseMultipartForm(w, r, 10, h.returns) {
		return
	}

	imageID := uuid.UUID(userUUID.Bytes).String()

	profileImagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		"pictures-profile", imageID, "image", h.logger, h.returns)
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

	h.returns.ReturnJSON(w, artists, http.StatusOK)
}
