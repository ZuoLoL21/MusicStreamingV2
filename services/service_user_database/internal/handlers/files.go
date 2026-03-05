package handlers

import (
	"backend/internal/di"
	"backend/internal/storage"
	"fmt"
	"io"
	"net/http"

	libsdi "libs/di"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// FileHandler handles file serving from storage
type FileHandler struct {
	fileStorage storage.FileStorageClient
	logger      *zap.Logger
	returns     *libsdi.ReturnManager
	config      *di.Config
	db          DB
}

// NewFileHandler creates a new file handler
func NewFileHandler(
	fileStorage storage.FileStorageClient,
	logger *zap.Logger,
	returns *libsdi.ReturnManager,
	config *di.Config,
	db DB,
) *FileHandler {
	return &FileHandler{
		fileStorage: fileStorage,
		logger:      logger,
		returns:     returns,
		config:      config,
		db:          db,
	}
}

// ServeFile serves a file from storage with authentication
func (h *FileHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	objectPath := r.URL.Path[len("/files/"):]

	resourceInfo, err := parseResourceFromPath(objectPath)
	if err != nil {
		h.logger.Warn("invalid file path", zap.String("path", objectPath), zap.Error(err))
		h.returns.ReturnError(w, "invalid file path", http.StatusBadRequest)
		return
	}

	// Extract user UUID
	userUUIDStr, _ := r.Context().Value(h.config.UserUUIDKey).(string)
	var userUUID pgtype.UUID
	if userUUIDStr != "" {
		userUUID, _ = uuidToPgtype(userUUIDStr)
	}

	// Check permissions
	allowed, err := checkFileAccess(r.Context(), h.db, resourceInfo, userUUID)
	if err != nil {
		h.logger.Error("failed to check file access",
			zap.String("path", objectPath),
			zap.Error(err))
		h.returns.ReturnError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !allowed {
		h.logger.Warn("unauthorized file access attempt",
			zap.String("path", objectPath),
			zap.String("user_uuid", userUUIDStr))
		h.returns.ReturnError(w, "forbidden", http.StatusForbidden)
		return
	}

	// Serve file
	object, contentType, size, err := h.fileStorage.GetObject(r.Context(), objectPath)
	if err != nil {
		h.logger.Warn("file not found", zap.String("path", objectPath), zap.Error(err))
		h.returns.ReturnError(w, "file not found", http.StatusNotFound)
		return
	}
	defer object.Close()

	// Set headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))

	if resourceInfo.IsDefault {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else if resourceInfo.ResourceType == "pictures-playlist" {
		w.Header().Set("Cache-Control", "private, max-age=3600, must-revalidate")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=604800")
	}

	// Stream file
	_, err = io.Copy(w, object)
	if err != nil {
		h.logger.Error("failed to stream file", zap.String("path", objectPath), zap.Error(err))
	}
}
