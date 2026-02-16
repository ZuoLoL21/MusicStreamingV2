package app

import (
	"file-storage/internal/di"
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
	Logger  *zap.Logger
	Config  *di.Config
	Storage *di.LocalStorageManager
	Returns *di.ReturnManager
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	musicHandler := handlers.NewMusicHandler(a.Logger, a.Config, a.Storage, a.Returns)
	imageHandler := handlers.NewImageHandler(a.Logger, a.Config, a.Storage, a.Returns)

	sMusic := r.PathPrefix("/music").Subrouter()
	sMusic.HandleFunc("/{id}", musicHandler.StreamAudio).Methods("GET")
	sMusic.HandleFunc("/{id}", musicHandler.SaveAudio).Methods("PUT")
	sMusic.HandleFunc("/{id}", musicHandler.DeleteAudio).Methods("DELETE")
	sMusic.HandleFunc("/{id}", musicHandler.UpdateAudio).Methods("POST")

	sImage := r.PathPrefix("/image").Subrouter()
	sImage.HandleFunc("/{folder}/{id}", imageHandler.GetImage).Methods("GET")
	sImage.HandleFunc("/{folder}/{id}", imageHandler.UpdateImage).Methods("POST")
	sImage.HandleFunc("/{folder}/", imageHandler.GetDefaultImage).Methods("GET")
	sImage.HandleFunc("/{folder}/{id}", imageHandler.DeleteImage).Methods("DELETE")

	r.HandleFunc("/", defaultEndpoint)

	r.Use(
		middleware.RequestIDMiddleware(a.Config),
		middleware.LoggingMiddleware(a.Logger, a.Config),
	)

	return r
}

func New(logger *zap.Logger, config *di.Config, storage *di.LocalStorageManager, returns *di.ReturnManager) *App {
	return &App{Logger: logger, Config: config, Storage: storage, Returns: returns}
}
