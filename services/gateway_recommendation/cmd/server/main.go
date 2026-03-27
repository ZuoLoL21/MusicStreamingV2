package main

import (
	"context"
	"errors"
	"gateway_recommendation/internal/app"
	"gateway_recommendation/internal/di"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	libsdi "libs/di"

	"go.uber.org/zap"
)

func initLogger(serviceName string) *zap.Logger {
	logger, _ := zap.NewProduction()
	logger = logger.WithOptions(zap.AddCaller())

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	logger = logger.With(
		zap.String("service_name", serviceName),
		zap.String("environment", environment),
	)

	return logger
}

func main() {
	// Logger
	logger := initLogger("gateway_recommendation")
	defer func() {
		_ = logger.Sync()
	}()

	// Init components
	config := di.LoadConfig(logger)
	jwtHandler := libsdi.GetJWTHandler(logger, config, config.ApplicationName)
	returnManager := libsdi.NewReturnManager(logger)

	// Router
	application := app.New(config, logger, jwtHandler, returnManager)
	srv := &http.Server{
		Handler:      application.Router(),
		Addr:         ":" + config.Port,
		WriteTimeout: 45 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in background
	go func() {
		logger.Info("Server listening", zap.String("port", config.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Server failed to start", zap.Error(err))
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
