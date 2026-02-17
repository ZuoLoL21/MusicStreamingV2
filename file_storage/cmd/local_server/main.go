package main

import (
	"context"
	"errors"
	"file-storage/internal/app"
	"file-storage/internal/di"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// Router
	application := app.New(logger, config, storage, returns)
	srv := &http.Server{
		Handler:      application.Router(),
		Addr:         config.ListenAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in background
	go func() {
		logger.Info("server starting", zap.String("addr", config.ListenAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Wait for SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	stop()

	// Graceful termination
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("server shutdown:", err)
	}
	logger.Info("server stopped")
}
