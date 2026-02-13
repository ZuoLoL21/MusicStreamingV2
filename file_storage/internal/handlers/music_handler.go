package handlers

import (
	"file-storage/internal/helpers"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func StreamAudio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseDir := helpers.GetDataFolder("music")
	validated := helpers.ValidateUUID(vars["id"])
	if !validated {
		http.Error(w, "Invalid id provided", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filepath.Join(baseDir, vars["id"]+".mp3"))
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

	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func SaveAudio(w http.ResponseWriter, r *http.Request) {
	response, err := helpers.ParseAudioFromRequest(r)

	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}
	id := response.ID
	part := response.Data

	baseDir := helpers.GetDataFolder("music")
	destPath := filepath.Join(baseDir, id+".mp3")

	writtenBytes, err := helpers.SaveToFile(part, destPath)
	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}

	if writtenBytes > helpers.MaxAudioSize {
		_ = os.Remove(destPath)
		http.Error(w, "audio exceeds maximum size", http.StatusRequestEntityTooLarge)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprintf(w, "Audio file %s saved successfully with (%d bytes)", id, writtenBytes)
}

func UpdateAudio(w http.ResponseWriter, r *http.Request) {
	response, err := helpers.ParseAudioFromRequest(r)

	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}
	id := response.ID
	part := response.Data

	baseDir := helpers.GetDataFolder("music")
	destPath := filepath.Join(baseDir, id+".mp3")

	// Check if file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		http.Error(w, "Audio file not found", http.StatusNotFound)
		return
	}

	writtenBytes, err := helpers.SaveToFile(part, destPath)
	if err != nil {
		http.Error(w, err.Message, err.Status)
		return
	}

	if writtenBytes > helpers.MaxAudioSize {
		_ = os.Remove(destPath)
		http.Error(w, "audio exceeds maximum size", http.StatusRequestEntityTooLarge)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Audio file %s updated successfully with (%d bytes)", id, writtenBytes)
}

func DeleteAudio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseDir := helpers.GetDataFolder("music")
	validated := helpers.ValidateUUID(vars["id"])
	if !validated {
		http.Error(w, "Invalid id provided", http.StatusBadRequest)
		return
	}

	destPath := filepath.Join(baseDir, vars["id"]+".mp3")

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		http.Error(w, "audio file not found", http.StatusNotFound)
		return
	}

	err := os.Remove(destPath)
	if err != nil {
		http.Error(w, "failed to delete audio file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Audio file %s deleted successfully", vars["id"])
}
