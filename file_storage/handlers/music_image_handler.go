package handlers

import (
	"fmt"
	"music-streaming/file-storage/helpers"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

const defaultImageName = "default_music.jpeg"

func GetMusicImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseDir := helpers.GetDataFolder("music_pictures")
	validated := helpers.ValidateUUID(vars["id"])
	if !validated {
		http.Error(w, "Invalid id provided", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filepath.Join(baseDir, vars["id"]+".jpeg"))
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Stat failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}
func GetDefaultMusicImage(w http.ResponseWriter, r *http.Request) {
	baseDir := helpers.GetDataFolder("default")
	file, err := os.Open(filepath.Join(baseDir, defaultImageName))
	if err != nil {
		http.Error(w, "Default not found", http.StatusNotFound)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Stat failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func UpdateMusicImage(w http.ResponseWriter, r *http.Request) {
	response, err := helpers.ParseImageFromRequest(r)

	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}
	id := response.ID
	part := response.Data

	baseDir := helpers.GetDataFolder("music_pictures")
	destPath := filepath.Join(baseDir, id+".jpeg")

	writtenBytes, err := helpers.SaveToFileB(part, destPath)
	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Music image %s saved successfully with (%d bytes)", id, writtenBytes)
}
