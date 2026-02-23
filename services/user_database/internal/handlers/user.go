package handlers

import (
	"fmt"
	"net/http"

	"backend/internal/di"
	sqlhandler "backend/sql/sqlc"

	libsdi "libs/di"
	libshelpers "libs/helpers"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type UserHandler struct {
	logger  *zap.Logger
	config  *di.Config
	secrets *di.SecretsManager
	returns *libsdi.ReturnManager
	db      DB
}

func NewUserHandler(logger *zap.Logger, config *di.Config, secrets *di.SecretsManager, returns *libsdi.ReturnManager, db DB) *UserHandler {
	return &UserHandler{
		logger:  logger,
		config:  config,
		secrets: secrets,
		returns: returns,
		db:      db,
	}
}

type tokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *UserHandler) issueTokenPair(uuidStr string) tokenPair {
	_, priKey, kid := h.secrets.GetKeyInfo(h.config.JWTStorePath)
	access := libshelpers.GenerateJwt(h.config.SubjectNormal, uuidStr, priKey, kid, h.config.JWTExpirationNormal)
	refresh := libshelpers.GenerateJwt(h.config.SubjectRefresh, uuidStr, priKey, kid, h.config.JWTExpirationRefresh)
	return tokenPair{AccessToken: access, RefreshToken: refresh}
}

type registerRequest struct {
	Username         string  `json:"username" validate:"required,min=5"`
	Email            string  `json:"email" validate:"required,email"`
	Password         string  `json:"password" validate:"required,min=8"`
	Bio              *string `json:"bio"`
	ProfileImagePath *string `json:"profile_image_path"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeBody[registerRequest](w, r, h.returns)
	if !ok {
		return
	}

	hashedPassword := libshelpers.Encode(body.Password)

	var bio pgtype.Text
	if body.Bio != nil {
		bio = pgtype.Text{String: *body.Bio, Valid: true}
	}

	var profileImagePath pgtype.Text
	if body.ProfileImagePath != nil {
		profileImagePath = pgtype.Text{String: *body.ProfileImagePath, Valid: true}
	}

	userUUID, err := h.db.CreateUser(r.Context(), sqlhandler.CreateUserParams{
		Username:         body.Username,
		Email:            body.Email,
		HashedPassword:   hashedPassword,
		Bio:              bio,
		ProfileImagePath: profileImagePath,
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

	uuidStr := uuid.UUID(userUUID.Bytes).String()
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
	uuidStr := r.Context().Value(h.config.UserUUIDKey).(string)
	_, priKey, kid := h.secrets.GetKeyInfo(h.config.JWTStorePath)
	access := libshelpers.GenerateJwt(h.config.SubjectNormal, uuidStr, priKey, kid, h.config.JWTExpirationNormal)
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

	var bio pgtype.Text
	if body.Bio != nil {
		bio = pgtype.Text{String: *body.Bio, Valid: true}
	}

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

type updateImageRequest struct {
	ProfileImagePath string `json:"profile_image_path" validate:"required"`
}

func (h *UserHandler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	body, ok := decodeBody[updateImageRequest](w, r, h.returns)
	if !ok {
		return
	}

	if err := h.db.UpdateImage(r.Context(), sqlhandler.UpdateImageParams{
		Uuid:             userUUID,
		ProfileImagePath: pgtype.Text{String: body.ProfileImagePath, Valid: true},
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
