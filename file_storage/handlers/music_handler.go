package handlers

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"music-streaming/file-storage/helpers"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func StreamAudio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseDir := helpers.Get_data_folder("music")
	file, err := os.Open(filepath.Join(baseDir, vars["id"]))
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()

	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func SaveAudio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"] + ".mp3"
	baseDir := helpers.Get_data_folder("music")

	// Use multipart reader for streaming (no memory limit)
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Failed to read multipart data", http.StatusBadRequest)
		return
	}

	// Get the first part (should be the audio file)
	part, err := reader.NextPart()
	if err != nil {
		http.Error(w, "Failed to get file part", http.StatusBadRequest)
		return
	}
	defer part.Close()
	if part.FormName() != "audio" {
		http.Error(w, "Expected 'audio' field", http.StatusBadRequest)
		return
	}

	err = testIfMP3(part)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the destination file
	destPath := filepath.Join(baseDir, id)
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// Stream directly to file (no memory buffer)
	written, err := io.Copy(destFile, part)
	if err != nil {
		err := os.Remove(destPath)
		if err != nil {
		}

		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Audio file %s saved successfully with (%d bytes)", id, written)
}

func UpdateAudio(w http.ResponseWriter, r *http.Request) {

}

func DeleteAudio(w http.ResponseWriter, r *http.Request) {

}

func testIfMP3(filePart *multipart.Part) error {
	header := make([]byte, 3)
	n, err := io.ReadFull(filePart, header)
	if err != nil || n != 3 {
		return errors.New("Failed to read file header")
	}
	isID3 := header[0] == 0x49 && header[1] == 0x44 && header[2] == 0x33
	isMPEG := header[0] == 0xFF && (header[1]&0xE0) == 0xE0
	if !isID3 && !isMPEG {
		return errors.New("Invalid MP3 file format")
	}
	return nil
}
