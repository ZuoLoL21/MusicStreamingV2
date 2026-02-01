package handlers

import (
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
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()

	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}
func SaveAudio(w http.ResponseWriter, r *http.Request) {

}
func UpdateAudio(w http.ResponseWriter, r *http.Request) {

}
func DeleteAudio(w http.ResponseWriter, r *http.Request) {

}
