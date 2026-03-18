package handlers

import (
	"backend/internal/consts"
	"backend/internal/storage"
	sqlhandler "backend/sql/sqlc"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	libsdi "libs/di"
	libsmiddleware "libs/middleware"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// parseMultipartForm parses multipart form with size limit and handles errors
func parseMultipartForm(w http.ResponseWriter, r *http.Request, maxSizeMB int64, returns *libsdi.ReturnManager) bool {
	logger := libsmiddleware.GetLogger(r.Context())

	maxSize := maxSizeMB << 20
	if err := r.ParseMultipartForm(maxSize); err != nil {
		logger.Warn("failed to parse multipart form",
			zap.Error(err),
			zap.String("content_type", r.Header.Get("Content-Type")))

		contentType := r.Header.Get("Content-Type")
		errorMsg := "request must be multipart/form-data"
		if contentType == "application/json" {
			errorMsg = "request must be multipart/form-data, not application/json"
		} else if contentType == "" {
			errorMsg = "missing Content-Type header, expected multipart/form-data"
		} else {
			errorMsg = fmt.Sprintf("invalid Content-Type: %s, expected multipart/form-data", contentType)
		}

		returns.ReturnError(w, errorMsg, http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

// handleFileStorageError maps file storage errors to appropriate HTTP responses
func handleFileStorageError(ctx context.Context, w http.ResponseWriter, err error, returns *libsdi.ReturnManager, operation string) {
	logger := libsmiddleware.GetLogger(ctx)

	logger.Error("file storage operation failed", zap.String("operation", operation), zap.Error(err))

	if strings.Contains(err.Error(), "status 413") {
		returns.ReturnError(w, "file size exceeds limit", http.StatusRequestEntityTooLarge)
	} else if strings.Contains(err.Error(), "status 400") {
		returns.ReturnError(w, "invalid file format", http.StatusBadRequest)
	} else {
		returns.ReturnError(w, "file storage service unavailable", http.StatusServiceUnavailable)
	}
}

// uploadImageFromForm uploads image from multipart form with validation, returns URL or handles error
func uploadImageFromForm(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	fileStorage storage.FileStorageClient,
	folder string,
	imageID string,
	formFieldName string,
	returns *libsdi.ReturnManager,
) (pgtype.Text, bool) {
	logger := libsmiddleware.GetLogger(ctx)

	// Image is optional
	imageFile, _, err := r.FormFile(formFieldName)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return pgtype.Text{}, true
		}
		logger.Warn("failed to read image file from form", zap.Error(err))
		returns.ReturnError(w, "failed to read image file", http.StatusBadRequest)
		return pgtype.Text{}, false
	}
	defer func(imageFile multipart.File) {
		_ = imageFile.Close()
	}(imageFile)

	// Validate and process image
	var config storage.ImageValidationConfig
	if folder == consts.PicturesProfileFolder || folder == consts.PicturesArtistFolder {
		config = storage.DefaultProfileImageConfig()
	} else {
		config = storage.DefaultMusicImageConfig()
	}

	processedData, _, err := storage.ValidateAndProcessImage(imageFile, config)
	if err != nil {
		logger.Warn("image validation failed",
			zap.String("folder", folder),
			zap.String("imageID", imageID),
			zap.Error(err))

		errorMsg := "invalid image"
		if strings.Contains(err.Error(), "file size") {
			errorMsg = fmt.Sprintf("image file too large (max %dMB)", config.MaxFileSizeMB)
		} else if strings.Contains(err.Error(), "must be square") {
			errorMsg = "image must be square (width must equal height)"
		} else if strings.Contains(err.Error(), "unsupported image format") {
			errorMsg = "unsupported image format (use JPEG, PNG, or WebP)"
		} else if strings.Contains(err.Error(), "invalid image format") {
			errorMsg = "invalid or corrupted image file"
		} else {
			errorMsg = err.Error()
		}

		returns.ReturnError(w, errorMsg, http.StatusBadRequest)
		return pgtype.Text{}, false
	}

	imageURL, err := fileStorage.SaveImage(ctx, folder, imageID, processedData)
	if err != nil {
		handleFileStorageError(ctx, w, err, returns, "upload image")
		return pgtype.Text{}, false
	}

	return pgtype.Text{String: imageURL, Valid: true}, true
}

// uploadAudioFromForm uploads audio from multipart form, handles errors
func uploadAudioFromForm(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	fileStorage storage.FileStorageClient,
	musicID string,
	formFieldName string,
	returns *libsdi.ReturnManager,
) (string, bool) {
	audioFile, _, err := r.FormFile(formFieldName)
	if err != nil {
		returns.ReturnError(w, "audio file is required", http.StatusBadRequest)
		return "", false
	}
	defer func(audioFile multipart.File) {
		_ = audioFile.Close()
	}(audioFile)

	audioURL, err := fileStorage.SaveAudio(ctx, musicID, audioFile)
	if err != nil {
		handleFileStorageError(ctx, w, err, returns, "upload audio")
		return "", false
	}

	return audioURL, true
}

// cleanupImage deletes uploaded image on rollback (logs warnings, doesn't fail)
func cleanupImage(ctx context.Context, fileStorage storage.FileStorageClient, folder, imageID string) {
	logger := libsmiddleware.GetLogger(ctx)

	if err := fileStorage.DeleteImage(ctx, folder, imageID); err != nil {
		logger.Warn("failed to clean up image after operation failure",
			zap.String("folder", folder),
			zap.String("imageID", imageID),
			zap.Error(err))
	}
}

// cleanupAudio deletes uploaded audio on rollback (logs warnings, doesn't fail)
func cleanupAudio(ctx context.Context, fileStorage storage.FileStorageClient, musicID string) {
	logger := libsmiddleware.GetLogger(ctx)

	if err := fileStorage.DeleteAudio(ctx, musicID); err != nil {
		logger.Warn("failed to clean up audio after operation failure",
			zap.String("musicID", musicID),
			zap.Error(err))
	}
}

// applyDefaultImageIfEmpty sets default image URL if field is empty based on entity type
// also transforms existing paths (folder/name.ext) to /files/ URLs (by using convertPathToFileURL)
// entityType should be one of: "user", "artist", "album", "playlist", "music"
func applyDefaultImageIfEmpty(imagePath *pgtype.Text, entityType string) {
	if imagePath == nil {
		return
	}

	if !imagePath.Valid || imagePath.String == "" {
		*imagePath = pgtype.Text{String: convertDefaultToFileURL(entityType), Valid: true}
	} else {
		*imagePath = pgtype.Text{String: convertPathToFileURL(imagePath.String), Valid: true}
	}
}

// optionalStringToPgtype converts *string to pgtype.Text
func optionalStringToPgtype(s *string) pgtype.Text {
	if s != nil && *s != "" {
		return pgtype.Text{String: *s, Valid: true}
	}
	return pgtype.Text{}
}

// ResourceInfo contains parsed information about a file resource
type ResourceInfo struct {
	ResourceType string // "audio", "pictures-playlist", etc.
	ResourceID   string // UUID of the resource
	IsDefault    bool   // true if "defaults/*"
}

// parseResourceFromPath extracts resource information from a storage path
func parseResourceFromPath(objectPath string) (*ResourceInfo, error) {
	parts := strings.SplitN(objectPath, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid object path format")
	}

	resourceType := parts[0]
	fileName := parts[1]

	info := &ResourceInfo{
		ResourceType: resourceType,
		IsDefault:    resourceType == consts.DefaultsFolder,
	}

	if info.IsDefault {
		info.ResourceID = fileName
		return info, nil
	}

	uuidStr := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	info.ResourceID = uuidStr

	return info, nil
}

// checkFileAccess verifies if a user has permission to access a file
func checkFileAccess(
	ctx context.Context,
	db consts.DB,
	resourceInfo *ResourceInfo,
	userUUID pgtype.UUID, // May be invalid if not authenticated
) (bool, error) {
	if resourceInfo.IsDefault {
		return true, nil
	}

	if resourceInfo.ResourceType != consts.PicturesPlaylistFolder {
		return true, nil
	}

	// For playlist images, check if playlist is public OR user is owner
	playlistUUID, err := uuidToPgtype(resourceInfo.ResourceID)
	if err != nil {
		return false, fmt.Errorf("invalid playlist UUID: %w", err)
	}
	allowed, err := db.IsPlaylistPublicOrOwnedByUser(ctx, sqlhandler.IsPlaylistPublicOrOwnedByUserParams{
		PlaylistUuid: playlistUUID,
		UserUuid:     userUUID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check playlist access: %w", err)
	}

	return allowed, nil
}

// convertPathToFileURL converts storage path to /files/public/ URL
func convertPathToFileURL(storagePath string) string {
	if storagePath == "" {
		return ""
	}

	if strings.HasPrefix(storagePath, consts.PicturesPlaylistFolder) {
		return consts.PrivatePathStart + storagePath
	}
	return consts.PublicPathStart + storagePath
}

// convertDefaultToFileURL returns /files/public/ URL for default images
func convertDefaultToFileURL(entityType string) string {
	var path string
	switch entityType {
	case "user":
		path = consts.DefaultProfileImagePath
	case "artist":
		path = consts.DefaultArtistImagePath
	case "album":
		path = consts.DefaultAlbumImagePath
	case "playlist":
		path = consts.DefaultPlaylistImagePath
	case "music":
		path = consts.DefaultMusicImagePath
	default:
		path = consts.DefaultMusicImagePath
	}
	return consts.PublicPathStart + path
}
