package app

import (
	"backend/internal/di"
	"backend/internal/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	logger  *zap.Logger
	config  *di.Config
	secrets *di.SecretsManager
	returns *di.ReturnManager
}

func New(logger *zap.Logger, config *di.Config, secrets *di.SecretsManager, returns *di.ReturnManager) *App {
	return &App{
		logger:  logger,
		config:  config,
		secrets: secrets,
		returns: returns,
	}
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	auth := middleware.NewAuthHandler(a.logger, a.config, a.secrets)

	r.Use(middleware.RequestIDMiddleware(a.config))
	r.Use(middleware.LoggingMiddleware(a.logger, a.config))
	r.Use(auth.GetAuthMiddleware())

	return r
}
