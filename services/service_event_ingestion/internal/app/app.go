package app

import (
	"event_ingestion/internal/di"
	"event_ingestion/internal/handlers"
	"libs/consts"

	libsdi "libs/di"
	libshandlers "libs/handlers"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	logger     *zap.Logger
	config     *di.Config
	jwtHandler *libsdi.JWTHandler
	returns    *libsdi.ReturnManager
	clickhouse *di.ClickHouseClient
	handlers   *HandlerRegistry
}

type HandlerRegistry struct {
	Event *handlers.EventHandler
}

func New(logger *zap.Logger, config *di.Config, jwtHandler *libsdi.JWTHandler, returns *libsdi.ReturnManager, clickhouse *di.ClickHouseClient) *App {
	return &App{
		logger:     logger,
		config:     config,
		jwtHandler: jwtHandler,
		returns:    returns,
		clickhouse: clickhouse,
	}
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	// Initialize all handlers
	a.initHandlers()

	// Setup middleware
	publicRouter, protectedRouter := a.setupMiddleware(r)

	// Register routes
	a.registerHealthRoutes(publicRouter)
	a.registerEventRoutes(protectedRouter)

	return r
}

func (a *App) initHandlers() {
	a.handlers = &HandlerRegistry{
		Event: handlers.NewEventHandler(a.config, a.returns, a.clickhouse),
	}
}

func (a *App) setupMiddleware(r *mux.Router) (*mux.Router, *mux.Router) {
	serviceAuthHandler := libsmiddleware.NewAuthHandler(
		a.logger,
		a.jwtHandler,
		a.returns,
		consts.JWTSubjectService,
	)

	publicRouter := r.PathPrefix("").Subrouter()
	protectedRouter := r.PathPrefix("").Subrouter()

	publicRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.FailureRecoveryMiddleware(a.logger),
		libsmiddleware.Logger(a.logger),
	)
	protectedRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.FailureRecoveryMiddleware(a.logger),
		serviceAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.logger),
	)

	return publicRouter, protectedRouter
}

func (a *App) registerHealthRoutes(r *mux.Router) {
	r.HandleFunc("/health", libshandlers.NewHealthCheckHandler("service-event-ingestion")).Methods("GET")
}

func (a *App) registerEventRoutes(r *mux.Router) {
	r.HandleFunc("/events/listen", a.handlers.Event.IngestListenEvent).Methods("POST")
	r.HandleFunc("/events/like", a.handlers.Event.IngestLikeEvent).Methods("POST")
	r.HandleFunc("/events/theme", a.handlers.Event.IngestThemeEvent).Methods("POST")
	r.HandleFunc("/events/user", a.handlers.Event.IngestUserDimEvent).Methods("POST")
}
