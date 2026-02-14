package handlers

import (
	"file-storage/internal/dependencies"
	"file-storage/internal/general"
	"file-storage/internal/helpers"
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
	config  *dependencies.Config
	storage dependencies.StorageHandler
}

func NewImageHandler(logger *zap.Logger, config *dependencies.Config, storage dependencies.StorageHandler) *ImageHandler {
	return &ImageHandler{logger: logger, config: config, storage: storage}
}

func (h *ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	vars := mux.Vars(r)
	bucketName := vars["folder"]
	id := vars["id"]

	baseDir, err_s := h.storage.GetDataFolder(bucketName)
	if err_s != nil {
		logger.Warn("Failed to get data from storage", zap.Error(err_s), zap.String("bucket", bucketName))
		http.Error(w, "invalid bucket name", http.StatusBadRequest)
		return
	}

	validated := general.ValidateUUID(id)
	if !validated {
		logger.Warn("Invalid uuid received", zap.String("bucket", bucketName), zap.String("uuid", id))
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filepath.Join(baseDir, id+".jpeg"))
	if err != nil {
		logger.Warn("Failed to open file", zap.Error(err), zap.String("bucket", bucketName), zap.String("id", id))
		http.Error(w, "failed to open file", http.StatusNotFound)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		logger.Warn("Failed to stat file", zap.Error(err), zap.String("bucket", bucketName), zap.String("id", id))
		http.Error(w, "failed to stat file", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}
func (h *ImageHandler) GetDefaultImage(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	vars := mux.Vars(r)
	bucketName := vars["folder"]

	baseDir, _ := h.storage.GetDataFolder("default")
	file, err := os.Open(filepath.Join(baseDir, defaultMap[bucketName]))
	if err != nil {
		logger.Warn("Default image not found", zap.String("bucket", bucketName), zap.Error(err))
		http.Error(w, "Default not found", http.StatusNotFound)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		logger.Warn("Failed to stat default image", zap.String("bucket", bucketName), zap.Error(err))
		http.Error(w, "Stat failed", http.StatusInternalServerError)
		return
	}

	logger.Info("Serving default image", zap.String("bucket", bucketName))
	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func (h *ImageHandler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	response, err := helpers.ParseImageFromRequest(r)
	if err != nil {
		logger.Warn("Failed to parse image from request", zap.Int("status", err.Status), zap.String("message", err.Message))
		http.Error(w, err.Message, err.Status)
		return
	}
	id := response.ID
	part := response.Data
	bucketName := response.Bucket

	baseDir, err_s := h.storage.GetDataFolder(bucketName)
	if err_s != nil {
		logger.Warn("Failed to get data from storage", zap.Error(err_s), zap.String("bucket", bucketName))
		http.Error(w, "invalid bucket name", http.StatusBadRequest)
		return
	}
	destPath := filepath.Join(baseDir, id+".jpeg")

	writtenBytes, err := h.storage.SaveToFileB(part, destPath)
	if err != nil {
		logger.Warn("Failed to save image file", zap.String("id", id), zap.Int("status", err.Status))
		http.Error(w, err.Message, err.Status)
		return
	}

	logger.Info("Image file saved", zap.String("id", id), zap.String("bucket", bucketName), zap.Int64("bytes", writtenBytes))
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Music image %s saved successfully with (%d bytes)", id, writtenBytes)
}
