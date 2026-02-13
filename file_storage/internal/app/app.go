package app

import (
	"file-storage/internal/handlers"
	"file-storage/internal/middleware"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func defaultEndpoint(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := fmt.Fprintf(w, "Invalid endpoint")
	if err != nil {
		panic(err)
	}
}

type App struct {
	Logger *zap.Logger
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	musicHandler := handlers.NewMusicHandler(a.Logger)
	imageHandler := handlers.NewImageHandler(a.Logger)

	sMusic := r.PathPrefix("/music").Subrouter()
	sMusic.HandleFunc("/{id}", musicHandler.StreamAudio).Methods("GET")
	sMusic.HandleFunc("/{id}", musicHandler.SaveAudio).Methods("PUT")
	sMusic.HandleFunc("/{id}", musicHandler.DeleteAudio).Methods("DELETE")
	sMusic.HandleFunc("/{id}", musicHandler.UpdateAudio).Methods("POST")

	sImage := r.PathPrefix("/image").Subrouter()
	sImage.HandleFunc("/{folder}/{id}", imageHandler.GetImage).Methods("GET")
	sImage.HandleFunc("/{folder}/{id}", imageHandler.UpdateImage).Methods("POST")
	sImage.HandleFunc("/{folder}/", imageHandler.GetDefaultImage).Methods("GET")

	r.HandleFunc("/", defaultEndpoint)

	r.Use(middleware.LoggingMiddleware(a.Logger))
	return r
}

func New(logger *zap.Logger) *App {
	return &App{Logger: logger}
}
