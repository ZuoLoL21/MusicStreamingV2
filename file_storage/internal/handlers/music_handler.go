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

const musicDir = "music"

type MusicHandler struct {
	logger  *zap.Logger
	config  *dependencies.Config
	storage dependencies.StorageHandler
}

func NewMusicHandler(logger *zap.Logger, config *dependencies.Config, storage dependencies.StorageHandler) *MusicHandler {
	return &MusicHandler{logger: logger, config: config, storage: storage}
}

func (h *MusicHandler) StreamAudio(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	vars := mux.Vars(r)
	baseDir, _ := h.storage.GetDataFolder(musicDir)

	validated := general.ValidateUUID(vars["id"])
	if !validated {
		logger.Warn("Invalid UUID provided", zap.String("id", vars["id"]))
		http.Error(w, "Invalid id provided", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filepath.Join(baseDir, vars["id"]+".mp3"))
	if err != nil {
		logger.Warn("File not found", zap.String("id", vars["id"]), zap.Error(err))
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stat, err := file.Stat()
	if err != nil {
		logger.Warn("Failed to stat file", zap.String("id", vars["id"]), zap.Error(err))
		http.Error(w, "Stat failed", http.StatusInternalServerError)
		return
	}

	logger.Info("Streaming audio", zap.String("id", vars["id"]))
	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func (h *MusicHandler) SaveAudio(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	response, err := helpers.ParseAudioFromRequest(r)
	if err != nil {
		logger.Warn("Failed to parse audio from request", zap.Int("status", err.Status), zap.String("message", err.Message))
		http.Error(w, err.Message, err.Status)
		return
	}
	id := response.ID
	part := response.Data

	baseDir, _ := h.storage.GetDataFolder(musicDir)
	destPath := filepath.Join(baseDir, id+".mp3")

	writtenBytes, err := h.storage.SaveToFile(part, destPath)
	if err != nil {
		logger.Warn("Failed to save audio file", zap.String("id", id), zap.Int("status", err.Status))
		http.Error(w, err.Message, err.Status)
		return
	}

	if writtenBytes > helpers.MaxAudioSize {
		_ = os.Remove(destPath)
		logger.Warn("Audio exceeds maximum size, deleted", zap.String("id", id), zap.Int64("bytes", writtenBytes))
		http.Error(w, "audio exceeds maximum size", http.StatusRequestEntityTooLarge)
		return
	}

	logger.Info("Audio file saved", zap.String("id", id), zap.Int64("bytes", writtenBytes))
	w.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprintf(w, "Audio file %s saved successfully with (%d bytes)", id, writtenBytes)
}

func (h *MusicHandler) UpdateAudio(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	response, err := helpers.ParseAudioFromRequest(r)
	if err != nil {
		logger.Warn("Failed to parse audio from request", zap.Int("status", err.Status), zap.String("message", err.Message))
		http.Error(w, err.Message, err.Status)
		return
	}
	id := response.ID
	part := response.Data

	baseDir, _ := h.storage.GetDataFolder(musicDir)
	destPath := filepath.Join(baseDir, id+".mp3")

	// Check if file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		logger.Warn("Audio file not found for update", zap.String("id", id))
		http.Error(w, "Audio file not found", http.StatusNotFound)
		return
	}

	writtenBytes, err := h.storage.SaveToFile(part, destPath)
	if err != nil {
		logger.Warn("Failed to save updated audio file", zap.String("id", id), zap.Int("status", err.Status))
		http.Error(w, err.Message, err.Status)
		return
	}

	if writtenBytes > helpers.MaxAudioSize {
		_ = os.Remove(destPath)
		logger.Warn("Updated audio exceeds maximum size, deleted", zap.String("id", id), zap.Int64("bytes", writtenBytes))
		http.Error(w, "audio exceeds maximum size", http.StatusRequestEntityTooLarge)
		return
	}

	logger.Info("Audio file updated", zap.String("id", id), zap.Int64("bytes", writtenBytes))
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Audio file %s updated successfully with (%d bytes)", id, writtenBytes)
}

func (h *MusicHandler) DeleteAudio(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	vars := mux.Vars(r)
	baseDir, _ := h.storage.GetDataFolder(musicDir)
	validated := general.ValidateUUID(vars["id"])
	if !validated {
		logger.Warn("Invalid UUID provided", zap.String("id", vars["id"]))
		http.Error(w, "Invalid id provided", http.StatusBadRequest)
		return
	}

	destPath := filepath.Join(baseDir, vars["id"]+".mp3")

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		logger.Warn("Audio file not found for deletion", zap.String("id", vars["id"]))
		http.Error(w, "audio file not found", http.StatusNotFound)
		return
	}

	err := os.Remove(destPath)
	if err != nil {
		logger.Warn("Failed to delete audio file", zap.String("id", vars["id"]), zap.Error(err))
		http.Error(w, "failed to delete audio file", http.StatusInternalServerError)
		return
	}

	logger.Info("Audio file deleted", zap.String("id", vars["id"]))
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Audio file %s deleted successfully", vars["id"])
}
