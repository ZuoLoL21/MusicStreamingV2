package handlers

import (
	"file-storage/internal/service"
	"net/http"
)

func GetProfileImage(w http.ResponseWriter, r *http.Request) {
	service.GetImage(w, r, "profile_pictures")
}
func GetDefaultProfileImage(w http.ResponseWriter, r *http.Request) {
	service.GetDefaultImage(w, r, "profile_pictures")
}
func UpdateProfileImage(w http.ResponseWriter, r *http.Request) {
	service.UpdateImage(w, r, "profile_pictures")
}
