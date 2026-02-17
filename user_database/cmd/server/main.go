package main

import (
	"backend/internal/app"
	"backend/internal/di"
	"context"
	"errors"
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
	defer func() {
		_ = logger.Sync()
	}()

	// Init components
	config := di.LoadConfig(logger)
	secrets := di.GetSecretsManager(logger, config)
	returns := di.GetReturnManager(logger, config)

	// Router
	application := app.New(logger, config, secrets, returns)
	srv := &http.Server{
		Handler:      application.Router(),
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in background
	go func() {
		logger.Info("server starting", zap.String("addr", ":8080"))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Graceful termination
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	stop()

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("server shutdown:", err)
	}
	logger.Info("server stopped")
}
