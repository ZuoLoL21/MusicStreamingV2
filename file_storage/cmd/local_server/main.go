package main

import (
	"file-storage/internal/app"
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
	config := app.LoadConfig(logger)
	storage := app.GetLocalStorageManager(logger, config)
	storage.InitStorage()

	// RESI API
	application := app.New(logger)
	srv := &http.Server{
		Handler:      application.Router(),
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Server is up and running!")
	log.Fatal(srv.ListenAndServe())
}
