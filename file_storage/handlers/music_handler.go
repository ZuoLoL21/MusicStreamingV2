package handlers

import (
	"net/http"
	"os"
)

func streamAudio(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("music/song.mp3")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()

	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func saveAudio(w http.ResponseWriter, r *http.Request) {

}

type MusicHandler struct{}

func (h *MusicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		streamAudio(w, r)
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodDelete:
	default:
	}

}
