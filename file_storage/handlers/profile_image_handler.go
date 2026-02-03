package handlers

import (
	"fmt"
	"io"
	"music-streaming/file-storage/helpers"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

const defaultProfileName = "default.jpeg"

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

func CreateDefaultProfileImage(w http.ResponseWriter, r *http.Request) {
	baseDirDefault := helpers.GetDataFolder("default")
	source, err := os.Open(filepath.Join(baseDirDefault, defaultProfileName))
	if err != nil {
		http.Error(w, "Default not found", http.StatusNotFound)
		return
	}
	defer func(source *os.File) {
		_ = source.Close()
	}(source)

	vars := mux.Vars(r)
	baseDir := helpers.GetDataFolder("profile_pictures")
	destinationDir := filepath.Join(baseDir, vars["id"]+".jpeg")
	if _, err := os.Stat(destinationDir); !os.IsNotExist(err) {
		http.Error(w, "Already exists, unable to save a default picture", http.StatusBadRequest)
		return
	}

	destination, err := os.Create(destinationDir)
	if err != nil {
		http.Error(w, "File not created", http.StatusInternalServerError)
		return
	}
	defer func(destination *os.File) {
		_ = destination.Close()
	}(destination)

	_, err = io.Copy(destination, source)
	if err != nil {
		http.Error(w, "File not copied", http.StatusInternalServerError)
		return
	}

	stat, _ := destination.Stat()

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), destination)

}
