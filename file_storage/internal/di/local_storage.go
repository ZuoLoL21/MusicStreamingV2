package di

import (
	"errors"
	"file-storage/internal/general"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"go.uber.org/zap"
)

var possibleStorages = []string{"default", "music", "music_pictures", "profile_pictures"}

type LocalStorageManager struct {
	config *Config
	logger *zap.Logger
}

func GetLocalStorageManager(logger *zap.Logger, config *Config) *LocalStorageManager {
	return &LocalStorageManager{logger: logger, config: config}
}

func (h *LocalStorageManager) InitStorage() {
	slogger := h.logger.With(
		zap.String("lifespan", "init"),
	).Sugar()

	slogger.Infof("Data location set to %v", h.config.StorageLocation)

	for _, storage := range possibleStorages {
		directory, _ := h.GetDataFolder(storage)
		err := os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			slogger.Panicf("Error creating directory: %v", err)
		}
	}

	h.logger = h.logger.With(
		zap.String("lifespan", "running"),
	)
}

func (h *LocalStorageManager) GetDataFolder(name string) (string, error) {
	if !slices.Contains(possibleStorages, name) {
		return "", errors.New("invalid storage name")
	}
	return filepath.Join(h.config.StorageLocation, name), nil
}

func (h *LocalStorageManager) SaveToFile(filePart io.Reader, location string) (int64, *general.ErrorResult) {
	// Create the destination file
	destFile, err := os.Create(location)
	if err != nil {
		h.logger.Warn("Error creating file",
			zap.String("location", location),
			zap.Error(err),
		)
		return 0, &general.ErrorResult{Message: "failed to create file", Status: http.StatusInternalServerError}
	}
	defer func(destFile *os.File) {
		_ = destFile.Close()
	}(destFile)

	// Stream directly to file
	written, err := io.Copy(destFile, filePart)
	if err != nil {
		_ = destFile.Close()
		_ = os.Remove(location)

		h.logger.Warn("Failed to save file",
			zap.String("location", location),
			zap.Error(err),
		)
		return 0, &general.ErrorResult{Message: "failed to save file", Status: http.StatusInternalServerError}
	}
	return written, nil
}

func (h *LocalStorageManager) SaveToFileB(filePart []byte, location string) (int64, *general.ErrorResult) {
	// Create the destination file
	destFile, err := os.Create(location)
	if err != nil {
		h.logger.Warn("Error creating file",
			zap.String("location", location),
			zap.Error(err),
		)
		return 0, &general.ErrorResult{Message: "failed to create file", Status: http.StatusInternalServerError}
	}
	defer func(destFile *os.File) {
		_ = destFile.Close()
	}(destFile)

	// Stream directly to file
	written, err := destFile.Write(filePart)
	if err != nil {
		_ = destFile.Close()
		_ = os.Remove(location)

		h.logger.Warn("Failed to save file",
			zap.String("location", location),
			zap.Error(err),
		)
		return 0, &general.ErrorResult{Message: "failed to save file", Status: http.StatusInternalServerError}
	}
	return int64(written), nil
}
