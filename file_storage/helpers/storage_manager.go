package helpers

import (
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

func SaveToFile(filePart *multipart.Part, location string) (int64, *ErrorResult) {
	// Create the destination file
	destFile, err := os.Create(location)
	if err != nil {
		return 0, &ErrorResult{Message: "failed to create file", Status: http.StatusInternalServerError}
	}
	defer func(destFile *os.File) {
		_ = destFile.Close()
	}(destFile)

	// Stream directly to file
	written, err := io.Copy(destFile, filePart)
	if err != nil {
		_ = destFile.Close()
		_ = os.Remove(location)

		return 0, &ErrorResult{Message: "failed to save file", Status: http.StatusInternalServerError}
	}
	return written, nil
}
