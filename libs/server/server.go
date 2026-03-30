package server

import (
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

// TimeoutConfig holds HTTP server timeout configuration
type TimeoutConfig struct {
	Write time.Duration
	Read  time.Duration
	Idle  time.Duration
}

// DefaultTimeouts returns standard timeout values
func DefaultTimeouts() TimeoutConfig {
	return TimeoutConfig{
		Write: 15 * time.Second,
		Read:  15 * time.Second,
		Idle:  120 * time.Second,
	}
}

// GatewayTimeouts returns timeout values suitable for gateway services
func GatewayTimeouts() TimeoutConfig {
	return TimeoutConfig{
		Write: 45 * time.Second,
		Read:  15 * time.Second,
		Idle:  120 * time.Second,
	}
}

// RunHTTPServer starts an HTTP server with graceful shutdown handling.
// It blocks until the server is shut down via SIGINT/SIGTERM.
func RunHTTPServer(logger *zap.Logger, addr string, handler http.Handler, timeouts TimeoutConfig) {
	srv := &http.Server{
		Handler:      handler,
		Addr:         addr,
		WriteTimeout: timeouts.Write,
		ReadTimeout:  timeouts.Read,
		IdleTimeout:  timeouts.Idle,
	}
	go func() {
		logger.Info("server starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server failed to start", zap.Error(err))
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
