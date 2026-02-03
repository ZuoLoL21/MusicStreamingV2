package helpers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func GetDataFolder(name string) string {
	dataDirectory := os.Getenv("DATA_LOCATION")
	filepath.Clean(dataDirectory)
	return filepath.Join(dataDirectory, name)
}

func SaveToFile(filePart io.Reader, location string) (int64, *ErrorResult) {
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

func SaveToFileB(filePart []byte, location string) (int64, *ErrorResult) {
	// Create the destination file
	destFile, err := os.Create(location)
	if err != nil {
		return 0, &ErrorResult{Message: "failed to create file", Status: http.StatusInternalServerError}
	}
	defer func(destFile *os.File) {
		_ = destFile.Close()
	}(destFile)

	// Stream directly to file
	written, err := destFile.Write(filePart)
	if err != nil {
		_ = destFile.Close()
		_ = os.Remove(location)

		return 0, &ErrorResult{Message: "failed to save file", Status: http.StatusInternalServerError}
	}
	return int64(written), nil
}
