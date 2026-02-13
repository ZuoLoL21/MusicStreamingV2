package main

import (
	"file-storage/internal/app"
	"file-storage/internal/helpers"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Logger
	logger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// Init components
	helpers.InitStorage()

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
