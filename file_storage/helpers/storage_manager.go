package helpers

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func GetDataFolder(name string) string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	projectRoot := cwd
	for {
		fileStoragePath := filepath.Join(projectRoot, "file_storage")
		if _, err := os.Stat(fileStoragePath); err == nil {
			break
		}

		// Move up one directory
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			break
		}
		projectRoot = parent
	}

	return filepath.Join(projectRoot, "file_storage", "data", name)
}

func SaveToFile(filePart *multipart.Part, location string) (int64, error, int) {
	// Create the destination file
	destFile, err := os.Create(location)
	if err != nil {
		return 0, errors.New("failed to create file"), http.StatusInternalServerError
	}
	defer destFile.Close()

	// Stream directly to file
	written, err := io.Copy(destFile, filePart)
	if err != nil {
		_ = destFile.Close()
		_ = os.Remove(location)

		return 0, errors.New("failed to save file"), http.StatusInternalServerError
	}
	return written, nil, http.StatusOK
}
