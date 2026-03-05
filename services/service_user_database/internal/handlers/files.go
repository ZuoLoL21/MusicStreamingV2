package handlers

import (
	"backend/internal/storage"
	"fmt"
	"io"
	"net/http"

	libsdi "libs/di"

	"go.uber.org/zap"
)

// FileHandler handles file serving from storage
type FileHandler struct {
	fileStorage storage.FileStorageClient
	logger      *zap.Logger
	returns     *libsdi.ReturnManager
}

// NewFileHandler creates a new file handler
func NewFileHandler(fileStorage storage.FileStorageClient, logger *zap.Logger, returns *libsdi.ReturnManager) *FileHandler {
	return &FileHandler{
		fileStorage: fileStorage,
		logger:      logger,
		returns:     returns,
	}
}

// ServeFile serves a file from storage
func (h *FileHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	objectPath := r.URL.Path[len("/files/"):]

	object, contentType, size, err := h.fileStorage.GetObject(r.Context(), objectPath)
	if err != nil {
		h.logger.Warn("file not found",
			zap.String("path", objectPath),
			zap.Error(err))
		h.returns.ReturnError(w, "file not found", http.StatusNotFound)
		return
	}
	defer object.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year

	// Stream the file to the response
	_, err = io.Copy(w, object)
	if err != nil {
		h.logger.Error("failed to stream file",
			zap.String("path", objectPath),
			zap.Error(err))
		return
	}
}
