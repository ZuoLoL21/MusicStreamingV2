package main

import (
	"backend/internal/app"
	"backend/internal/client"
	"backend/internal/di"
	sqlhandler "backend/sql/sqlc"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	libsdi "libs/di"

	"github.com/jackc/pgx/v5/pgxpool"
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
	logger := initLogger("service_user_database")
	defer func() {
		_ = logger.Sync()
	}()

	// Init components
	config := di.LoadConfig(logger)
	jwtHandler := libsdi.GetJWTHandler(logger, config, config.ApplicationName)
	returns := libsdi.NewReturnManager(logger)

	// Database
	pool, err := pgxpool.New(context.Background(), config.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to create connection pool", zap.Error(err))
	}
	defer pool.Close()
	db := sqlhandler.New(pool)

	// File Storage Client
	fileStorage, err := client.NewMinIOFileStorageClient(
		config.MinIOEndpoint,
		config.MinIOAccessKey,
		config.MinIOSecretKey,
		config.MinIOBucketName,
		logger,
	)
	if err != nil {
		logger.Fatal("failed to create MinIO client", zap.Error(err))
	}

	// Router
	application := app.New(logger, config, jwtHandler, returns, db, fileStorage)
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
