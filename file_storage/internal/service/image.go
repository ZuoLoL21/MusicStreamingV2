package service

import (
	"file-storage/internal/helpers"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

var defaultMap = map[string]string{
	"music_pictures":   "default_music.jpeg",
	"profile_pictures": "default_profile.jpeg",
}

func GetImage(w http.ResponseWriter, r *http.Request, bucketName string) {
	vars := mux.Vars(r)
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
func GetDefaultImage(w http.ResponseWriter, r *http.Request, bucketName string) {
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

func UpdateImage(w http.ResponseWriter, r *http.Request, bucketName string) {
	response, err := helpers.ParseImageFromRequest(r)

	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}
	id := response.ID
	part := response.Data

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
