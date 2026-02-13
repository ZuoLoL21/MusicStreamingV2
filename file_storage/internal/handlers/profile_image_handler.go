package handlers

import (
	"file-storage/internal/helpers"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

const defaultProfileName = "default_profile.jpeg"

func GetProfileImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseDir := helpers.GetDataFolder("profile_pictures")
	details, err := helpers.RetrieveImage(vars["id"], baseDir)
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
func GetDefaultProfileImage(w http.ResponseWriter, r *http.Request) {
	details, err := helpers.RetrieveDefaultImage(defaultProfileName)
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
