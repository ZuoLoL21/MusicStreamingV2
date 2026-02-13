package main

import (
	"file-storage/internal/app"
	"file-storage/internal/dependencies"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func main() {
	// Logger
	logger, _ := zap.NewProduction()
	logger = logger.WithOptions(zap.AddCaller())
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	// Init components
	config := dependencies.LoadConfig(logger)
	storage := dependencies.GetLocalStorageManager(logger, config)
	storage.InitStorage()

	// RESI API
	application := app.New(logger, config, storage)
	srv := &http.Server{
		Handler:      application.Router(),
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Server is up and running!")
	log.Fatal(srv.ListenAndServe())
}
