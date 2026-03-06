package handlers

import (
	"backend/internal/consts"
	"backend/internal/di"
	"backend/internal/storage"
	"fmt"
	"io"
	libsconsts "libs/consts"
	"net/http"
	"strings"

	libsdi "libs/di"

	"go.uber.org/zap"
)

// FileHandler handles file serving from storage
type FileHandler struct {
	fileStorage storage.FileStorageClient
	logger      *zap.Logger
	returns     *libsdi.ReturnManager
	config      *di.Config
	db          consts.DB
}

// NewFileHandler creates a new file handler
func NewFileHandler(
	fileStorage storage.FileStorageClient,
	logger *zap.Logger,
	returns *libsdi.ReturnManager,
	config *di.Config,
	db consts.DB,
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
	objectPath := r.URL.Path

	// File info
	isPrivate := false
	var actualPath string

	if strings.HasPrefix(objectPath, consts.PrivatePathStart) {
		isPrivate = true
		actualPath = strings.TrimPrefix(objectPath, consts.PrivatePathStart)
	} else if strings.HasPrefix(objectPath, consts.PublicPathStart) {
		actualPath = strings.TrimPrefix(objectPath, consts.PublicPathStart)
	} else {
		h.logger.Warn("invalid path", zap.String("path", r.URL.Path))
		h.returns.ReturnError(w, "not valid path: must use public or private", http.StatusBadRequest)
		return
	}

	resourceInfo, err := parseResourceFromPath(actualPath)
	if err != nil {
		h.logger.Warn("invalid file path", zap.String("path", objectPath), zap.Error(err))
		h.returns.ReturnError(w, "invalid file path", http.StatusBadRequest)
		return
	}

	// For private files, require authentication and check permissions
	if isPrivate {
		userUUIDStr, ok := r.Context().Value(libsconsts.UserUUIDKey).(string)
		if !ok || userUUIDStr == "" {
			h.logger.Warn("unauthenticated access to private file",
				zap.String("path", objectPath))
			h.returns.ReturnError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userUUID, _ := uuidToPgtype(userUUIDStr)

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
	}

	// Serve file using the actual storage path (without public/private prefix)
	object, contentType, size, err := h.fileStorage.GetObject(r.Context(), actualPath)
	if err != nil {
		h.logger.Warn("file not found", zap.String("path", objectPath), zap.Error(err))
		h.returns.ReturnError(w, "file not found", http.StatusNotFound)
		return
	}
	defer func(object io.ReadCloser) {
		_ = object.Close()
	}(object)

	// Set headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))

	if resourceInfo.IsDefault {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else if resourceInfo.ResourceType == consts.PicturesPlaylistFolder {
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
