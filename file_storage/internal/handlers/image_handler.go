package handlers

import (
	"file-storage/internal/di"
	"file-storage/internal/general"
	"file-storage/internal/service"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var defaultMap = map[string]string{
	"music_pictures":   "default_music.jpeg",
	"profile_pictures": "default_profile.jpeg",
}

type ImageHandler struct {
	logger  *zap.Logger
	config  *di.Config
	storage *di.LocalStorageManager
	returns *di.ReturnManager
}

func NewImageHandler(logger *zap.Logger, config *di.Config, storage *di.LocalStorageManager, returns *di.ReturnManager) *ImageHandler {
	return &ImageHandler{logger: logger, config: config, storage: storage, returns: returns}
}

func (h *ImageHandler) loggerFor(r *http.Request) *zap.Logger {
	requestID, _ := r.Context().Value(h.config.RequestIDKey).(string)
	return h.logger.With(
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)
}

func (h *ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	logger := h.loggerFor(r)

	vars := mux.Vars(r)
	bucketName := vars["folder"]
	id := vars["id"]

	baseDir, errS := h.storage.GetDataFolder(bucketName)
	if errS != nil {
		logger.Info("Invalid bucket name", zap.Error(errS), zap.String("bucket", bucketName))
		h.returns.ReturnError(w, "invalid bucket name", http.StatusBadRequest)
		return
	}

	if !general.ValidateUUID(id) {
		logger.Info("Invalid uuid received", zap.String("bucket", bucketName), zap.String("uuid", id))
		h.returns.ReturnError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filepath.Join(baseDir, id+".jpeg"))
	if err != nil {
		logger.Info("Didn't find file", zap.Error(err), zap.String("bucket", bucketName), zap.String("id", id))
		h.returns.ReturnError(w, "didn't find file", http.StatusNotFound)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		logger.Warn("failed to stat file", zap.Error(err), zap.String("bucket", bucketName), zap.String("id", id))
		h.returns.ReturnError(w, "failed to stat file", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnFile(w, r, stat.Name(), stat.ModTime(), file)
}

func (h *ImageHandler) GetDefaultImage(w http.ResponseWriter, r *http.Request) {
	logger := h.loggerFor(r)

	vars := mux.Vars(r)
	bucketName := vars["folder"]

	baseDir, _ := h.storage.GetDataFolder("default")
	imageName := defaultMap[bucketName]
	if imageName == "" {
		logger.Warn("invalid bucket name", zap.String("bucket", bucketName))
		h.returns.ReturnError(w, "invalid bucket name", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filepath.Join(baseDir, imageName))
	if err != nil {
		logger.Warn("default image not found", zap.String("bucket", bucketName), zap.Error(err), zap.String("file_name", imageName), zap.String("folder", baseDir))
		h.returns.ReturnError(w, "default not found", http.StatusInternalServerError)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		logger.Warn("failed to stat default image", zap.String("bucket", bucketName), zap.Error(err))
		h.returns.ReturnError(w, "stat failed", http.StatusInternalServerError)
		return
	}

	logger.Info("serving default image", zap.String("bucket", bucketName))
	h.returns.ReturnFile(w, r, stat.Name(), stat.ModTime(), file)
}

func (h *ImageHandler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	logger := h.loggerFor(r)

	vars := mux.Vars(r)
	id := vars["id"]
	bucketName := vars["folder"]

	if !general.ValidateUUID(id) {
		logger.Info("invalid UUID provided", zap.String("id", id))
		h.returns.ReturnError(w, "invalid id provided", http.StatusBadRequest)
		return
	}

	baseDir, errS := h.storage.GetDataFolder(bucketName)
	if errS != nil {
		logger.Info("invalid bucket name", zap.Error(errS), zap.String("bucket", bucketName))
		h.returns.ReturnError(w, "invalid bucket name", http.StatusBadRequest)
		return
	}

	response, err := service.ParseImageFromRequest(r, id, bucketName)
	if err != nil {
		logger.Info("failed to parse image from request", zap.Int("status", err.Status), zap.String("message", err.Message))
		h.returns.ReturnError(w, err.Message, err.Status)
		return
	}

	destPath := filepath.Join(baseDir, id+".jpeg")

	writtenBytes, err := h.storage.SaveToFileB(response.Data, destPath)
	if err != nil {
		logger.Warn("failed to save image file", zap.String("id", id), zap.Int("status", err.Status))
		h.returns.ReturnError(w, err.Message, err.Status)
		return
	}

	logger.Info("image file saved", zap.String("id", id), zap.String("bucket", bucketName), zap.Int64("bytes", writtenBytes))
	h.returns.ReturnText(w, fmt.Sprintf("music image %s saved successfully with (%d bytes)", id, writtenBytes), http.StatusOK)
}
