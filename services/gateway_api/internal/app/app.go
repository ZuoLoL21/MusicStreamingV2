package app

import (
	"gateway_api/internal/di"
	"gateway_api/internal/handlers"
	"libs/consts"
	"libs/metrics"

	libsclients "libs/clients"
	libsdi "libs/di"
	libshandlers "libs/handlers"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	config               *di.Config
	logger               *zap.Logger
	jwtHandler           *libsdi.JWTHandler
	returns              *libsdi.ReturnManager
	userDBClient         *libsclients.ProxyClient
	recommendClient      *libsclients.ProxyClient
	eventIngestionClient *libsclients.ProxyClient
	handlers             *HandlerRegistry
}

type HandlerRegistry struct {
	Proxy *handlers.ProxyHandler
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()
	a.initHandlers()

	normalRouter := r.PathPrefix("").Subrouter()
	publicRouter, refreshRouter, protectedRouter := a.setupMiddleware(normalRouter)

	// Register routes
	a.registerMonitoringRoutes(r)
	a.registerPublicRoutes(publicRouter)
	a.registerRefreshRoutes(refreshRouter)
	a.registerProtectedRoutes(protectedRouter)

	return r
}

func (a *App) initHandlers() {
	a.handlers = &HandlerRegistry{
		Proxy: handlers.NewProxyHandler(
			a.userDBClient,
			a.recommendClient,
			a.eventIngestionClient,
		),
	}
}

func (a *App) setupMiddleware(normalRouter *mux.Router) (*mux.Router, *mux.Router, *mux.Router) {
	// Auth middleware setup
	normalAuthHandler := libsmiddleware.NewAuthHandler(
		a.logger,
		a.jwtHandler,
		a.returns,
		consts.JWTSubjectNormal,
	)
	refreshAuthHandler := libsmiddleware.NewAuthHandler(
		a.logger,
		a.jwtHandler,
		a.returns,
		consts.JWTSubjectRefresh,
	)
	serviceJWTHandler := libsmiddleware.NewServiceJWTHandler(
		a.logger,
		a.jwtHandler,
		a.returns,
		a.config.JWTExpirationService,
	)

	// Setting up route middleware
	publicRouter := normalRouter.PathPrefix("").Subrouter()
	refreshRouter := normalRouter.PathPrefix("").Subrouter()
	protectedRouter := normalRouter.PathPrefix("").Subrouter()

	normalRouter.Use(
		libsmiddleware.CORSMiddleware,
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.MetricsMiddleware(a.logger),
		libsmiddleware.FailureRecoveryMiddleware(a.logger),
	)
	publicRouter.Use(
		libsmiddleware.Logger(a.logger),
	)
	refreshRouter.Use(
		refreshAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.logger),
		serviceJWTHandler.GetServiceJWTMiddleware(),
	)
	protectedRouter.Use(
		normalAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.logger),
		serviceJWTHandler.GetServiceJWTMiddleware(),
	)

	return publicRouter, refreshRouter, protectedRouter
}

func (a *App) registerMonitoringRoutes(r *mux.Router) {
	r.Handle("/metrics", metrics.Handler()).Methods("GET")
	r.HandleFunc("/health", libshandlers.NewHealthCheckHandler("gateway-api")).Methods("GET")
}

func (a *App) registerPublicRoutes(r *mux.Router) {
	r.HandleFunc("/login", a.handlers.Proxy.ProxyLogin).Methods("POST", "PUT", "OPTIONS")
	r.PathPrefix("/files/public/").HandlerFunc(a.handlers.Proxy.ProxyPublicFiles).Methods("GET")
}

func (a *App) registerRefreshRoutes(r *mux.Router) {
	r.HandleFunc("/renew", a.handlers.Proxy.ProxyRenew).Methods("POST")
}

func (a *App) registerProtectedRoutes(r *mux.Router) {
	// File routes
	r.PathPrefix("/files/private/").HandlerFunc(a.handlers.Proxy.ProxyPrivateFiles).Methods("GET")

	// User Database Service routes
	r.PathPrefix("/users").HandlerFunc(a.handlers.Proxy.ProxyUserDatabase)
	r.PathPrefix("/artists").HandlerFunc(a.handlers.Proxy.ProxyUserDatabase)
	r.PathPrefix("/albums").HandlerFunc(a.handlers.Proxy.ProxyUserDatabase)
	r.PathPrefix("/music").HandlerFunc(a.handlers.Proxy.ProxyUserDatabase)
	r.PathPrefix("/tags").HandlerFunc(a.handlers.Proxy.ProxyUserDatabase)
	r.PathPrefix("/playlists").HandlerFunc(a.handlers.Proxy.ProxyUserDatabase)
	r.PathPrefix("/history").HandlerFunc(a.handlers.Proxy.ProxyUserDatabase)
	r.PathPrefix("/search").HandlerFunc(a.handlers.Proxy.ProxyUserDatabase)

	// Recommendation Service routes
	r.PathPrefix("/recommend").HandlerFunc(a.handlers.Proxy.ProxyRecommendation)
	r.PathPrefix("/popular").HandlerFunc(a.handlers.Proxy.ProxyRecommendation)

	// Event Ingestion Service routes
	r.PathPrefix("/events").HandlerFunc(a.handlers.Proxy.ProxyEventIngestion)
}

func New(
	config *di.Config,
	logger *zap.Logger,
	jwtHandler *libsdi.JWTHandler,
	returns *libsdi.ReturnManager,
) *App {
	userDBClient := libsclients.NewProxyClient(config.UserDatabaseServiceURL)
	recommendClient := libsclients.NewProxyClient(config.RecommendationServiceURL)
	eventIngestionClient := libsclients.NewProxyClient(config.EventIngestionServiceURL)

	return &App{
		config:               config,
		logger:               logger,
		jwtHandler:           jwtHandler,
		returns:              returns,
		userDBClient:         userDBClient,
		recommendClient:      recommendClient,
		eventIngestionClient: eventIngestionClient,
	}
}
