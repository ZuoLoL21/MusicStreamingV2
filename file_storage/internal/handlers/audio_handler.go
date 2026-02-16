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

const musicDir = "music"

type MusicHandler struct {
	logger  *zap.Logger
	config  *dependencies.Config
	storage *dependencies.LocalStorageManager
	returns *dependencies.ReturnManager
}

func NewMusicHandler(logger *zap.Logger, config *dependencies.Config, storage *dependencies.LocalStorageManager, returns *dependencies.ReturnManager) *MusicHandler {
	return &MusicHandler{logger: logger, config: config, storage: storage, returns: returns}
}

func (h *MusicHandler) StreamAudio(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	vars := mux.Vars(r)
	id := vars["id"]
	baseDir, _ := h.storage.GetDataFolder(musicDir)

	if !general.ValidateUUID(id) {
		logger.Info("invalid UUID provided", zap.String("id", id))
		h.returns.ReturnError(w, r, "invalid UUID provided", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filepath.Join(baseDir, id+".mp3"))
	if err != nil {
		logger.Info("file not found", zap.String("id", id), zap.Error(err))
		h.returns.ReturnError(w, r, "file not found", http.StatusNotFound)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stat, err := file.Stat()
	if err != nil {
		logger.Warn("failed to stat file", zap.String("id", id), zap.Error(err))
		h.returns.ReturnError(w, r, "stat failed", http.StatusInternalServerError)
		return
	}

	logger.Info("streaming audio", zap.String("id", id))
	h.returns.ReturnFile(w, r, stat.Name(), stat.ModTime(), file)
}

func (h *MusicHandler) SaveAudio(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	vars := mux.Vars(r)
	id := vars["id"]
	if !general.ValidateUUID(id) {
		logger.Info("invalid UUID provided", zap.String("id", id))
		h.returns.ReturnError(w, r, "invalid id provided", http.StatusBadRequest)
		return
	}

	response, err := service.ParseAudioFromRequest(r, id)
	if err != nil {
		logger.Warn("failed to parse audio from request", zap.Int("status", err.Status), zap.String("message", err.Message))
		h.returns.ReturnError(w, r, err.Message, err.Status)
		return
	}

	baseDir, _ := h.storage.GetDataFolder(musicDir)
	destPath := filepath.Join(baseDir, id+".mp3")

	writtenBytes, err := h.storage.SaveToFile(response.Data, destPath)
	if err != nil {
		logger.Warn("failed to save audio file", zap.String("id", id), zap.Int("status", err.Status))
		h.returns.ReturnError(w, r, err.Message, err.Status)
		return
	}

	if writtenBytes > service.MaxAudioSize {
		_ = os.Remove(destPath)
		logger.Info("audio exceeds maximum size, deleted", zap.String("id", id), zap.Int64("bytes", writtenBytes))
		h.returns.ReturnError(w, r, "audio exceeds maximum size", http.StatusRequestEntityTooLarge)
		return
	}

	logger.Info("audio file saved", zap.String("id", id), zap.Int64("bytes", writtenBytes))
	h.returns.ReturnText(
		w,
		fmt.Sprintf("audio file %s saved successfully with (%d bytes)", id, writtenBytes),
		http.StatusCreated,
	)
}

func (h *MusicHandler) UpdateAudio(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	vars := mux.Vars(r)
	id := vars["id"]
	if !general.ValidateUUID(id) {
		logger.Info("invalid UUID provided", zap.String("id", id))
		h.returns.ReturnError(w, r, "invalid id provided", http.StatusBadRequest)
		return
	}

	response, err := service.ParseAudioFromRequest(r, id)
	if err != nil {
		logger.Warn("failed to parse audio from request", zap.Int("status", err.Status), zap.String("message", err.Message))
		h.returns.ReturnError(w, r, err.Message, err.Status)
		return
	}

	baseDir, _ := h.storage.GetDataFolder(musicDir)
	destPath := filepath.Join(baseDir, id+".mp3")

	// Check if file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		logger.Info("audio file not found for update", zap.String("id", id))
		h.returns.ReturnError(w, r, "audio file not found for update", http.StatusNotFound)
		return
	}

	writtenBytes, err := h.storage.SaveToFile(response.Data, destPath)
	if err != nil {
		logger.Warn("failed to save updated audio file", zap.String("id", id), zap.Int("status", err.Status))
		h.returns.ReturnError(w, r, err.Message, err.Status)
		return
	}

	if writtenBytes > service.MaxAudioSize {
		_ = os.Remove(destPath)
		logger.Info("updated audio exceeds maximum size, deleted", zap.String("id", id), zap.Int64("bytes", writtenBytes))
		h.returns.ReturnError(w, r, "audio exceeds maximum size", http.StatusRequestEntityTooLarge)
		return
	}

	logger.Info("audio file updated", zap.String("id", id), zap.Int64("bytes", writtenBytes))
	h.returns.ReturnText(
		w,
		fmt.Sprintf("audio file %s updated successfully with (%d bytes)", id, writtenBytes),
		http.StatusOK,
	)
}

func (h *MusicHandler) DeleteAudio(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	vars := mux.Vars(r)
	id := vars["id"]
	baseDir, _ := h.storage.GetDataFolder(musicDir)

	if !general.ValidateUUID(id) {
		logger.Warn("invalid UUID provided", zap.String("id", id))
		h.returns.ReturnError(w, r, "invalid id provided", http.StatusBadRequest)
		return
	}

	destPath := filepath.Join(baseDir, id+".mp3")

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		logger.Warn("audio file not found for deletion", zap.String("id", id))
		h.returns.ReturnError(w, r, "audio file not found", http.StatusNotFound)
		return
	}

	err := os.Remove(destPath)
	if err != nil {
		logger.Warn("failed to delete audio file", zap.String("id", id), zap.Error(err))
		h.returns.ReturnError(w, r, "failed to delete audio file", http.StatusInternalServerError)
		return
	}

	logger.Info("audio file deleted", zap.String("id", id))
	h.returns.ReturnText(
		w,
		fmt.Sprintf("audio file %s deleted successfully", id),
		http.StatusOK,
	)
}
