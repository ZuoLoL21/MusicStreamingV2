package handlers

import (
	"backend/internal/client"
	"backend/internal/consts"
	"backend/internal/di"
	"backend/internal/storage"
	sqlhandler "backend/sql/sqlc"
	"context"
	"errors"
	"fmt"
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

type AuthHandler struct {
	config         *di.Config
	jwtHandler     *libsdi.JWTHandler
	returns        *libsdi.ReturnManager
	db             consts.DB
	fileStorage    storage.FileStorageClient
	clickhouseSync *client.ClickHouseSync
}

func NewAuthHandler(config *di.Config, jwtHandler *libsdi.JWTHandler, returns *libsdi.ReturnManager, db consts.DB, fileStorage storage.FileStorageClient, clickhouseSync *client.ClickHouseSync) *AuthHandler {
	return &AuthHandler{
		config:         config,
		jwtHandler:     jwtHandler,
		returns:        returns,
		db:             db,
		fileStorage:    fileStorage,
		clickhouseSync: clickhouseSync,
	}
}

type tokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserUUID     string `json:"user_uuid"`
	DeviceID     string `json:"device_id"`
}

// issueTokenPair generates access and refresh tokens, stores the hashed refresh token in the database
//
// Use empty for deviceName if you don't wish to change the database
func (h *AuthHandler) issueTokenPair(uuidStr, deviceID, deviceName string) (tokenPair, error) {
	access, err := h.jwtHandler.GenerateJwtWithDevice(
		libsconsts.JWTSubjectNormal,
		uuidStr,
		deviceID,
		h.config.JWTExpirationNormal,
	)
	if err != nil {
		return tokenPair{}, err
	}

	refresh, err := h.jwtHandler.GenerateJwtWithDevice(
		libsconsts.JWTSubjectRefresh,
		uuidStr,
		deviceID,
		h.config.JWTExpirationRefresh,
	)
	if err != nil {
		return tokenPair{}, err
	}

	// Store token hash
	userUUID, deviceUUID, err := convertUUIDs(uuidStr, deviceID)
	if err != nil {
		return tokenPair{}, err
	}

	tokenHash := libshelpers.HashToken(refresh)
	_, err = h.db.UpsertRefreshToken(context.Background(), sqlhandler.UpsertRefreshTokenParams{
		UserUuid:   userUUID,
		DeviceID:   deviceUUID,
		TokenHash:  tokenHash,
		DeviceName: optionalStringToPgtext(deviceName),
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(h.config.JWTExpirationRefresh), Valid: true},
	})
	if err != nil {
		return tokenPair{}, err
	}

	return tokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		UserUUID:     uuidStr,
		DeviceID:     deviceID,
	}, nil
}

type loginRequest struct {
	Email      string  `json:"email" validate:"required,email"`
	Password   string  `json:"password" validate:"required"`
	DeviceID   string  `json:"device_id" validate:"required,uuid"`
	DeviceName *string `json:"device_name"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
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
	deviceName := optionalPointerToString(body.DeviceName)
	tokens, err := h.issueTokenPair(uuidStr, body.DeviceID, deviceName)
	if err != nil {
		logger.Error("failed to generate tokens", zap.Error(err))
		h.returns.ReturnError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("token pair issued",
		zap.String("user_uuid", uuidStr),
		zap.String("device_id", body.DeviceID),
		zap.String("operation", "login"))
	h.returns.ReturnJSON(w, tokens, http.StatusOK)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	// Ensure is multipart form
	if !parseMultipartForm(w, r, 10, h.returns) {
		return
	}

	// Retrieval + validation
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

	deviceID := r.FormValue("device_id")
	if deviceID == "" {
		h.returns.ReturnError(w, "device_id required", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(deviceID); err != nil {
		h.returns.ReturnError(w, "invalid device_id format (must be UUID)", http.StatusBadRequest)
		return
	}

	deviceName := r.FormValue("device_name")

	hashedPassword := libshelpers.Encode(password)
	bioText := optionalStringToPgtype(bio)

	// Creation!
	createdUUID, err := h.db.CreateUser(r.Context(), sqlhandler.CreateUserParams{
		Username:       username,
		Email:          email,
		HashedPassword: hashedPassword,
		Bio:            bioText,
		Country:        country,
	})
	if err != nil {
		logger.Warn("failed to create user", zap.Error(err))

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
		consts.PicturesProfileFolder, userID, "image", h.returns)
	if !ok {
		return
	}

	// Update
	if err := h.db.UpdateImage(r.Context(), sqlhandler.UpdateImageParams{
		Uuid:             createdUUID,
		ProfileImagePath: profileImagePath,
	}); err != nil {
		logger.Error("failed to update user image after creation", zap.Error(err))
		if profileImagePath.Valid {
			cleanupImage(r.Context(), h.fileStorage, consts.PicturesProfileFolder, userID)
		}
		logger.Warn("user created but image update failed - image cleaned up",
			zap.String("userID", userID),
			zap.Error(err))
	}

	h.clickhouseSync.SyncUserDim(createdUUID, deviceID, country, time.Now())

	// Issue
	uuidStr := uuid.UUID(createdUUID.Bytes).String()
	tokens, err := h.issueTokenPair(uuidStr, deviceID, deviceName)
	if err != nil {
		logger.Error("failed to generate tokens", zap.Error(err))
		h.returns.ReturnError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	h.returns.ReturnJSON(w, tokens, http.StatusCreated)
}

func (h *AuthHandler) Renew(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	uuidStr := libshelpers.GetUserUUIDFromContext(r.Context())
	deviceID := libshelpers.GetDeviceIDFromContext(r.Context())

	if uuidStr == "" || deviceID == "" {
		logger.Error("missing user UUID or device ID in context")
		h.returns.ReturnError(w, "invalid token", http.StatusUnauthorized)
		return
	}

	err := h.validateRenewToken(r, uuidStr, deviceID)
	if err != nil {
		logger.Warn("failed to get renew token", zap.Error(err))
		h.returns.ReturnError(w, "invalid token", http.StatusUnauthorized)
		return
	}

	tokens, err := h.issueTokenPair(uuidStr, deviceID, "")
	if err != nil {
		logger.Error("failed to generate tokens", zap.Error(err))
		h.returns.ReturnError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("token pair renewed",
		zap.String("user_uuid", uuidStr),
		zap.String("device_id", deviceID))

	h.returns.ReturnJSON(w, tokens, http.StatusOK)
}

func (h *AuthHandler) getRenewToken(r *http.Request) (string, error) {
	refreshTokenHeader := r.Header.Get("X-Refresh-Token")
	if refreshTokenHeader == "" {
		return "", fmt.Errorf("missing X-Refresh-Token header")
	}
	refreshToken := strings.TrimPrefix(refreshTokenHeader, "Bearer ")
	if refreshToken == "" {
		return "", fmt.Errorf("empty refresh token after trimming")
	}

	return refreshToken, nil
}

func (h *AuthHandler) validateRenewToken(r *http.Request, uuidStr, deviceID string) error {
	logger := libsmiddleware.GetLogger(r.Context())

	refreshToken, err := h.getRenewToken(r)
	if err != nil {
		return err
	}

	tokenHash := libshelpers.HashToken(refreshToken)
	storedToken, err := h.db.GetRefreshTokenByHash(r.Context(), tokenHash)
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}
	if !verifyTokenMatch(storedToken, uuidStr, deviceID, logger) {
		return fmt.Errorf("invalid token")
	}

	return nil
}
