package handlers

import (
	"file-storage/internal/dependencies"
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
	config  *dependencies.Config
	storage dependencies.LocalStorageManager
}

func NewImageHandler(logger *zap.Logger, config *dependencies.Config, storage dependencies.LocalStorageManager) *ImageHandler {
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

	baseDir, errS := h.storage.GetDataFolder(bucketName)
	if errS != nil {
		logger.Info("Invalid bucket name", zap.Error(errS), zap.String("bucket", bucketName))
		http.Error(w, "invalid bucket name", http.StatusBadRequest)
		return
	}

	validated := general.ValidateUUID(id)
	if !validated {
		logger.Info("Invalid uuid received", zap.String("bucket", bucketName), zap.String("uuid", id))
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filepath.Join(baseDir, id+".jpeg"))
	if err != nil {
		logger.Info("Didn't find file", zap.Error(err), zap.String("bucket", bucketName), zap.String("id", id))
		http.Error(w, "didn't find file", http.StatusNotFound)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		logger.Warn("failed to stat file", zap.Error(err), zap.String("bucket", bucketName), zap.String("id", id))
		http.Error(w, "failed to stat file", http.StatusInternalServerError)
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
		logger.Warn("default image not found", zap.String("bucket", bucketName), zap.Error(err), zap.String("file_name", defaultMap[bucketName]), zap.String("folder", baseDir))
		http.Error(w, "default not found", http.StatusInternalServerError)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		logger.Warn("failed to stat default image", zap.String("bucket", bucketName), zap.Error(err))
		http.Error(w, "stat failed", http.StatusInternalServerError)
		return
	}

	logger.Info("serving default image", zap.String("bucket", bucketName))
	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func (h *ImageHandler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	response, err := service.ParseImageFromRequest(r)
	if err != nil {
		logger.Info("failed to parse image from request", zap.Int("status", err.Status), zap.String("message", err.Message))
		http.Error(w, err.Message, err.Status)
		return
	}
	id := response.ID
	part := response.Data
	bucketName := response.Bucket

	baseDir, errS := h.storage.GetDataFolder(bucketName)
	if errS != nil {
		logger.Info("invalid bucket name", zap.Error(errS), zap.String("bucket", bucketName))
		http.Error(w, "invalid bucket name", http.StatusBadRequest)
		return
	}
	destPath := filepath.Join(baseDir, id+".jpeg")

	writtenBytes, err := h.storage.SaveToFileB(part, destPath)
	if err != nil {
		logger.Warn("failed to save image file", zap.String("id", id), zap.Int("status", err.Status))
		http.Error(w, err.Message, err.Status)
		return
	}

	logger.Info("image file saved", zap.String("id", id), zap.String("bucket", bucketName), zap.Int64("bytes", writtenBytes))
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "music image %s saved successfully with (%d bytes)", id, writtenBytes)
}
