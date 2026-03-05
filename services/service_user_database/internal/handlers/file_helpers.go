package handlers

import (
	"backend/internal/storage"
	"context"
	"fmt"
	"net/http"
	"strings"

	libsdi "libs/di"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// parseMultipartForm parses multipart form with size limit and handles errors
func parseMultipartForm(w http.ResponseWriter, r *http.Request, maxSizeMB int64, returns *libsdi.ReturnManager, logger *zap.Logger) bool {
	maxSize := maxSizeMB << 20
	if err := r.ParseMultipartForm(maxSize); err != nil {
		logger.Warn("failed to parse multipart form",
			zap.Error(err),
			zap.String("content_type", r.Header.Get("Content-Type")),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))

		contentType := r.Header.Get("Content-Type")
		errorMsg := "request must be multipart/form-data"
		if contentType == "application/json" {
			errorMsg = "request must be multipart/form-data, not application/json"
		} else if contentType == "" {
			errorMsg = "missing Content-Type header, expected multipart/form-data"
		} else {
			errorMsg = fmt.Sprintf("invalid Content-Type: %s, expected multipart/form-data", contentType)
		}

		returns.ReturnError(w, errorMsg, http.StatusBadRequest)
		return false
	}
	return true
}

// handleFileStorageError maps file storage errors to appropriate HTTP responses
func handleFileStorageError(w http.ResponseWriter, err error, logger *zap.Logger, returns *libsdi.ReturnManager, operation string) {
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
	logger *zap.Logger,
	returns *libsdi.ReturnManager,
) (pgtype.Text, bool) {
	imageFile, _, err := r.FormFile(formFieldName)
	if err != nil {
		// Image is optional
		if err == http.ErrMissingFile {
			return pgtype.Text{}, true
		}
		returns.ReturnError(w, "failed to read image file", http.StatusBadRequest)
		return pgtype.Text{}, false
	}
	defer imageFile.Close()

	// Validate and process image
	var config storage.ImageValidationConfig
	if folder == "pictures-profile" || folder == "pictures-artist" {
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
		handleFileStorageError(w, err, logger, returns, "upload image")
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
	logger *zap.Logger,
	returns *libsdi.ReturnManager,
) (string, bool) {
	audioFile, _, err := r.FormFile(formFieldName)
	if err != nil {
		returns.ReturnError(w, "audio file is required", http.StatusBadRequest)
		return "", false
	}
	defer audioFile.Close()

	audioURL, err := fileStorage.SaveAudio(ctx, musicID, audioFile)
	if err != nil {
		handleFileStorageError(w, err, logger, returns, "upload audio")
		return "", false
	}

	return audioURL, true
}

// cleanupImage deletes uploaded image on rollback (logs warnings, doesn't fail)
func cleanupImage(ctx context.Context, fileStorage storage.FileStorageClient, folder, imageID string, logger *zap.Logger) {
	if err := fileStorage.DeleteImage(ctx, folder, imageID); err != nil {
		logger.Warn("failed to clean up image after operation failure",
			zap.String("folder", folder),
			zap.String("imageID", imageID),
			zap.Error(err))
	}
}

// cleanupAudio deletes uploaded audio on rollback (logs warnings, doesn't fail)
func cleanupAudio(ctx context.Context, fileStorage storage.FileStorageClient, musicID string, logger *zap.Logger) {
	if err := fileStorage.DeleteAudio(ctx, musicID); err != nil {
		logger.Warn("failed to clean up audio after operation failure",
			zap.String("musicID", musicID),
			zap.Error(err))
	}
}

// applyDefaultImageIfEmpty sets default image URL if field is empty based on entity type
// entityType should be one of: "user", "artist", "album", "playlist", "music"
// Also transforms existing paths (folder/name.ext) to public URLs
func applyDefaultImageIfEmpty(imagePath *pgtype.Text, fileStorage storage.FileStorageClient, entityType string) {
	if imagePath == nil {
		return
	}

	if !imagePath.Valid || imagePath.String == "" {
		var defaultURL string
		switch entityType {
		case "user":
			defaultURL = fileStorage.GetDefaultProfileImageURL()
		case "artist":
			defaultURL = fileStorage.GetDefaultArtistImageURL()
		case "album":
			defaultURL = fileStorage.GetDefaultAlbumImageURL()
		case "playlist":
			defaultURL = fileStorage.GetDefaultPlaylistImageURL()
		case "music":
			defaultURL = fileStorage.GetDefaultMusicImageURL()
		default:
			panic("invalid entity type")
		}
		*imagePath = pgtype.Text{String: defaultURL, Valid: true}
	} else {
		*imagePath = pgtype.Text{String: fileStorage.BuildPublicURL(imagePath.String), Valid: true}
	}
}

// optionalStringToPgtype converts *string to pgtype.Text
func optionalStringToPgtype(s *string) pgtype.Text {
	if s != nil && *s != "" {
		return pgtype.Text{String: *s, Valid: true}
	}
	return pgtype.Text{}
}
