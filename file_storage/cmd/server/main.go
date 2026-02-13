package main

import (
	"file-storage/internal/handlers"
	"file-storage/internal/helpers"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func defaultEndpoint(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := fmt.Fprintf(w, "Hello World")
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("Data location set to " + helpers.GetDataFolder(""))

	r := mux.NewRouter()

	sMusic := r.PathPrefix("/music").Subrouter()
	sMusic.HandleFunc("/{id}", handlers.StreamAudio).Methods("GET")
	sMusic.HandleFunc("/{id}", handlers.SaveAudio).Methods("PUT")
	sMusic.HandleFunc("/{id}", handlers.DeleteAudio).Methods("DELETE")
	sMusic.HandleFunc("/{id}", handlers.UpdateAudio).Methods("POST")

	sProfile := r.PathPrefix("/profile").Subrouter()
	sProfile.HandleFunc("/{id}", handlers.GetProfileImage).Methods("GET")
	sProfile.HandleFunc("/{id}", handlers.UpdateProfileImage).Methods("POST")
	sProfile.HandleFunc("/", handlers.GetDefaultProfileImage).Methods("GET")

	sMusicImage := r.PathPrefix("/music/image").Subrouter()
	sMusicImage.HandleFunc("/{id}", handlers.GetMusicImage).Methods("GET")
	sMusicImage.HandleFunc("/{id}", handlers.UpdateMusicImage).Methods("POST")
	sMusicImage.HandleFunc("/", handlers.GetDefaultMusicImage).Methods("GET")

	r.HandleFunc("/", defaultEndpoint)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Server is up and running!")
	log.Fatal(srv.ListenAndServe())
}
