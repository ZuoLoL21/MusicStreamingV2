package app

import (
	"event_ingestion/internal/di"
	"event_ingestion/internal/handlers"
	"libs/consts"
	"libs/metrics"

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

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()
	a.initHandlers()

	normalRouter := r.PathPrefix("").Subrouter()
	protectedRouter := a.setupMiddleware(normalRouter)

	// Register routes
	a.registerMonitoringRoutes(r)
	a.registerEventRoutes(protectedRouter)

	return r
}

func (a *App) initHandlers() {
	a.handlers = &HandlerRegistry{
		Event: handlers.NewEventHandler(a.config, a.returns, a.clickhouse),
	}
}

func (a *App) setupMiddleware(normalRouter *mux.Router) *mux.Router {
	// Auth middleware setup
	serviceAuthHandler := libsmiddleware.NewAuthHandler(
		a.logger,
		a.jwtHandler,
		a.returns,
		consts.JWTSubjectService,
	)

	// Setting up route middleware
	protectedRouter := normalRouter.PathPrefix("").Subrouter()

	normalRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.MetricsMiddleware(a.logger),
		libsmiddleware.FailureRecoveryMiddleware(a.logger),
	)
	protectedRouter.Use(
		serviceAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.logger),
	)

	return protectedRouter
}

func (a *App) registerMonitoringRoutes(r *mux.Router) {
	r.Handle("/metrics", metrics.Handler()).Methods("GET")
	r.HandleFunc("/health", libshandlers.NewHealthCheckHandler("service-event-ingestion")).Methods("GET")
}

func (a *App) registerEventRoutes(r *mux.Router) {
	r.HandleFunc("/events/listen", a.handlers.Event.IngestListenEvent).Methods("POST")
	r.HandleFunc("/events/like", a.handlers.Event.IngestLikeEvent).Methods("POST")
	r.HandleFunc("/events/theme", a.handlers.Event.IngestThemeEvent).Methods("POST")
	r.HandleFunc("/events/user", a.handlers.Event.IngestUserDimEvent).Methods("POST")
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
