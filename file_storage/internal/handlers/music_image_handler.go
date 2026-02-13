package handlers

import (
	"file-storage/internal/service"
	"net/http"
)

func GetMusicImage(w http.ResponseWriter, r *http.Request) {
	service.GetImage(w, r, "music_pictures")
}
func GetDefaultMusicImage(w http.ResponseWriter, r *http.Request) {
	service.GetDefaultImage(w, r, "music_pictures")
}
func UpdateMusicImage(w http.ResponseWriter, r *http.Request) {
	service.UpdateImage(w, r, "music_pictures")
}
