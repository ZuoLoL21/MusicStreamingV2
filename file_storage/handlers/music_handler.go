package handlers

import (
	"fmt"
	"mime/multipart"
	"music-streaming/file-storage/helpers"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func StreamAudio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseDir := helpers.GetDataFolder("music")
	file, err := os.Open(filepath.Join(baseDir, vars["id"]))
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	stat, _ := file.Stat()

	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func SaveAudio(w http.ResponseWriter, r *http.Request) {
	id, part, err, errCode := helpers.ParseAudioFromRequest(r)

	if err != nil {
		http.Error(w, err.Error(), errCode)
		return
	}
	defer func(part *multipart.Part) {
		_ = part.Close()
	}(part)

	baseDir := helpers.GetDataFolder("music")
	destPath := filepath.Join(baseDir, id)

	writtenBytes, err, errCode := helpers.SaveToFile(part, destPath)
	if err != nil {
		http.Error(w, err.Error(), errCode)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprintf(w, "Audio file %s saved successfully with (%d bytes)", id, writtenBytes)
}

func UpdateAudio(w http.ResponseWriter, r *http.Request) {

}

func DeleteAudio(w http.ResponseWriter, r *http.Request) {

}
