package main

import (
	"file-storage/internal/app"
	"file-storage/internal/di"
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
	config := di.LoadConfig(logger)
	storage := di.GetLocalStorageManager(logger, config)
	storage.InitStorage()
	returns := di.GetReturnManager(logger, config)

	// RESI API
	application := app.New(logger, config, storage, returns)
	srv := &http.Server{
		Handler:      application.Router(),
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Server is up and running!")
	log.Fatal(srv.ListenAndServe())
}
