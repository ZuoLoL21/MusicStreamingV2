package handlers

import (
	"backend/internal/consts"
	"backend/internal/di"
	"backend/internal/storage"
	sqlhandler "backend/sql/sqlc"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	libsconsts "libs/consts"
	"net/http"
	"strings"
	"time"

	libsdi "libs/di"
	libshelpers "libs/helpers"
	libsmiddleware "libs/middleware"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type UserHandler struct {
	logger      *zap.Logger
	config      *di.Config
	jwtHandler  *libsdi.JWTHandler
	returns     *libsdi.ReturnManager
	db          consts.DB
	fileStorage storage.FileStorageClient
	httpClient  *http.Client
}

func NewUserHandler(logger *zap.Logger, config *di.Config, jwtHandler *libsdi.JWTHandler, returns *libsdi.ReturnManager, db consts.DB, fileStorage storage.FileStorageClient) *UserHandler {
	return &UserHandler{
		logger:      logger,
		config:      config,
		jwtHandler:  jwtHandler,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

type tokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserUUID     string `json:"user_uuid"`
}

func (h *UserHandler) issueTokenPair(uuidStr string) (tokenPair, error) {
	access, err := h.jwtHandler.GenerateJwt(
		libsconsts.JWTSubjectNormal,
		uuidStr,
		h.config.JWTExpirationNormal,
	)
	if err != nil {
		return tokenPair{}, err
	}

	refresh, err := h.jwtHandler.GenerateJwt(
		libsconsts.JWTSubjectRefresh,
		uuidStr,
		h.config.JWTExpirationRefresh,
	)
	if err != nil {
		return tokenPair{}, err
	}

	return tokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		UserUUID:     uuidStr,
	}, nil
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns, h.logger) {
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
	if !isValidEmail(email) {
		h.returns.ReturnError(w, "invalid email format", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	if password == "" {
		h.returns.ReturnError(w, "password required", http.StatusBadRequest)
		return
	}
	if !isValidPassword(password) {
		h.returns.ReturnError(w, "password must be at least 8 characters and contain uppercase, lowercase, number, and special character", http.StatusBadRequest)
		return
	}

	country := r.FormValue("country")
	if country == "" || len(country) != 2 {
		h.returns.ReturnError(w, "country required (ISO 3166-1 alpha-2 code)", http.StatusBadRequest)
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
		Country:        country,
	})
	if err != nil {
		h.logger.Warn("failed to create user", zap.Error(err))

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			var message string
			if strings.Contains(pgErr.ConstraintName, "email") {
				message = "email already in use"
			} else if strings.Contains(pgErr.ConstraintName, "username") {
				message = "username already taken"
			} else {
				message = "duplicate entry"
			}
			h.returns.ReturnError(w, message, http.StatusConflict)
			return
		}

		h.returns.ReturnError(
			w,
			"failed to create user",
			http.StatusInternalServerError,
		)
		return
	}

	userID := uuid.UUID(createdUUID.Bytes).String()

	// Upload
	profileImagePath, ok := uploadImageFromForm(r.Context(), w, r, h.fileStorage,
		consts.PicturesProfileFolder, userID, "image", h.logger, h.returns)
	if !ok {
		return
	}

	// Update
	if err := h.db.UpdateImage(r.Context(), sqlhandler.UpdateImageParams{
		Uuid:             createdUUID,
		ProfileImagePath: profileImagePath,
	}); err != nil {
		h.logger.Error("failed to update user image after creation", zap.Error(err))
		if profileImagePath.Valid {
			cleanupImage(r.Context(), h.fileStorage, consts.PicturesProfileFolder, userID, h.logger)
		}
		h.logger.Warn("user created but image update failed - image cleaned up",
			zap.String("userID", userID),
			zap.Error(err))
	}

	go h.syncUserDimToClickHouse(createdUUID, country, time.Now())

	uuidStr := uuid.UUID(createdUUID.Bytes).String()
	tokens, err := h.issueTokenPair(uuidStr)
	if err != nil {
		h.logger.Error("failed to generate tokens", zap.Error(err))
		h.returns.ReturnError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	h.returns.ReturnJSON(w, tokens, http.StatusCreated)
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	body, ok := decodeBody[loginRequest](w, r, h.returns)
	if !ok {
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		logger.Warn("login attempt for non-existent user",
			zap.String("attempted_email", body.Email))
		h.returns.ReturnError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !libshelpers.Verify(body.Password, user.HashedPassword) {
		logger.Warn("login attempt with invalid password",
			zap.String("user_uuid", uuidToString(user.Uuid)))
		h.returns.ReturnError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	uuidStr := uuid.UUID(user.Uuid.Bytes).String()
	tokens, err := h.issueTokenPair(uuidStr)
	if err != nil {
		logger.Error("failed to generate tokens", zap.Error(err))
		h.returns.ReturnError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	logger.Info("user logged in successfully",
		zap.String("user_uuid", uuidStr))
	h.logger.Info("token pair issued",
		zap.String("user_uuid", uuidStr),
		zap.String("operation", "login"))
	h.returns.ReturnJSON(w, tokens, http.StatusOK)
}

func (h *UserHandler) Renew(w http.ResponseWriter, r *http.Request) {
	uuidStr := libshelpers.GetUserUUIDFromContext(r.Context())
	if uuidStr == "" {
		h.returns.ReturnError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	access, err := h.jwtHandler.GenerateJwt(
		libsconsts.JWTSubjectNormal,
		uuidStr,
		h.config.JWTExpirationNormal,
	)
	if err != nil {
		h.logger.Error("failed to generate access token", zap.Error(err))
		h.returns.ReturnError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("access token renewed",
		zap.String("user_uuid", uuidStr))

	h.returns.ReturnJSON(w, map[string]string{"access_token": access}, http.StatusOK)
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

	go h.syncUserDimToClickHouse(userUUID, body.Country, time.Now())

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

// syncUserDimToClickHouse sends user dimension data to the event ingestion service
func (h *UserHandler) syncUserDimToClickHouse(userUUID pgtype.UUID, country string, createdAt time.Time) {
	userUUIDStr := uuid.UUID(userUUID.Bytes).String()
	if h.config.EventIngestionServiceURL == "" {
		return
	}

	payload := map[string]interface{}{
		"user_uuid":  userUUIDStr,
		"created_at": createdAt.Format(time.RFC3339),
		"country":    country,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("failed to marshal user dim payload", zap.Error(err))
		return
	}

	req, err := http.NewRequest("POST", h.config.EventIngestionServiceURL+"/events/user", bytes.NewBuffer(jsonData))
	if err != nil {
		h.logger.Error("failed to create user dim request", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json")

	serviceJWT, err := h.jwtHandler.GenerateJwt(
		libsconsts.JWTSubjectService,
		userUUIDStr,
		h.config.JWTExpirationService,
	)
	if err != nil {
		h.logger.Error("failed to generate service JWT", zap.Error(err))
		return
	}
	req.Header.Set("Authorization", "Bearer "+serviceJWT)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Warn("failed to sync user dim to ClickHouse", zap.Error(err))
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		h.logger.Warn("user dim sync failed", zap.Int("status", resp.StatusCode))
	} else {
		h.logger.Debug("user dim synced to ClickHouse", zap.String("user_uuid", userUUIDStr))
	}
}
