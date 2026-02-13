package handlers

import (
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
	logger *zap.Logger
}

func NewImageHandler(logger *zap.Logger) *ImageHandler {
	return &ImageHandler{logger: logger}
}

func (*ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["folder"]

	baseDir := helpers.GetDataFolder(bucketName)

	details, err := helpers.RetrieveImage(vars["id"], baseDir)
	if err != nil {
		http.Error(w, err.Error(), err.Status)
		return
	}
	file := details.File
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, details.Name, details.ModTime, file)
}
func (*ImageHandler) GetDefaultImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["folder"]

	details, err := helpers.RetrieveDefaultImage(defaultMap[bucketName])
	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}
	file := details.File
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, details.Name, details.ModTime, file)
}

func (*ImageHandler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	response, err := helpers.ParseImageFromRequest(r)

	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}
	id := response.ID
	part := response.Data
	bucketName := response.Bucket

	baseDir := helpers.GetDataFolder(bucketName)
	destPath := filepath.Join(baseDir, id+".jpeg")

	writtenBytes, err := helpers.SaveToFileB(part, destPath)
	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Music image %s saved successfully with (%d bytes)", id, writtenBytes)
}
