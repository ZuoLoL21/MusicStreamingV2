package handlers

import "net/http"

func getProfile(w http.ResponseWriter, r *http.Request)    {}
func updateProfile(w http.ResponseWriter, r *http.Request) {}

type ProfileHandler struct{}

func (h *ProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getProfile(w, r)
	case http.MethodPost:
		updateProfile(w, r)
	case http.MethodPut:
		updateProfile(w, r)

	default:
	}

}
