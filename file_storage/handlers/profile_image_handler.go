package handlers

import (
	"fmt"
	"music-streaming/file-storage/helpers"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

const defaultProfileName = "default_profile.jpeg"

func GetProfileImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseDir := helpers.GetDataFolder("profile_pictures")
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
func GetDefaultProfileImage(w http.ResponseWriter, r *http.Request) {
	baseDir := helpers.GetDataFolder("default")
	file, err := os.Open(filepath.Join(baseDir, defaultProfileName))
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

func UpdateProfileImage(w http.ResponseWriter, r *http.Request) {
	response, err := helpers.ParseImageFromRequest(r)

	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}
	id := response.ID
	part := response.Data

	baseDir := helpers.GetDataFolder("profile_pictures")
	destPath := filepath.Join(baseDir, id+".jpeg")

	writtenBytes, err := helpers.SaveToFileB(part, destPath)
	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Profile picture %s saved successfully with (%d bytes)", id, writtenBytes)
}
