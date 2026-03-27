package app

import (
	"gateway_api/internal/clients"
	"gateway_api/internal/di"
	"gateway_api/internal/handlers"
	"libs/consts"
	"libs/metrics"

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
	userDBClient         *clients.UserDatabaseClient
	recommendClient      *clients.RecommendationClient
	eventIngestionClient *clients.EventIngestionClient
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	r.Use(libsmiddleware.CORSMiddleware)

	// Create handlers
	proxyHandler := handlers.NewProxyHandler(
		a.userDBClient,
		a.recommendClient,
		a.eventIngestionClient,
	)
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

	publicRouter := r.PathPrefix("").Subrouter()
	refreshRouter := r.PathPrefix("").Subrouter()
	protectedRouter := r.PathPrefix("").Subrouter()

	publicRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.MetricsMiddleware(a.logger),
		libsmiddleware.FailureRecoveryMiddleware(a.logger),
		libsmiddleware.Logger(a.logger),
	)
	refreshRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.MetricsMiddleware(a.logger),
		libsmiddleware.FailureRecoveryMiddleware(a.logger),
		refreshAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.logger),
		serviceJWTHandler.GetServiceJWTMiddleware(),
	)
	protectedRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.MetricsMiddleware(a.logger),
		libsmiddleware.FailureRecoveryMiddleware(a.logger),
		normalAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.logger),
		serviceJWTHandler.GetServiceJWTMiddleware(),
	)

	// Public routes
	r.Handle("/metrics", metrics.Handler()).Methods("GET")
	publicRouter.HandleFunc("/health", libshandlers.NewHealthCheckHandler("gateway-api")).Methods("GET")
	publicRouter.HandleFunc("/login", proxyHandler.ProxyLogin).Methods("POST", "PUT", "OPTIONS")
	publicRouter.PathPrefix("/files/public/").HandlerFunc(proxyHandler.ProxyPublicFiles).Methods("GET")

	// Renewal
	refreshRouter.HandleFunc("/renew", proxyHandler.ProxyRenew).Methods("POST")

	// Protected
	protectedRouter.PathPrefix("/files/private/").HandlerFunc(proxyHandler.ProxyPrivateFiles).Methods("GET")

	// User Database Service routes
	protectedRouter.PathPrefix("/users").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/artists").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/albums").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/music").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/tags").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/playlists").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/history").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/search").HandlerFunc(proxyHandler.ProxyUserDatabase)

	// Recommendation Service routes
	protectedRouter.PathPrefix("/recommend").HandlerFunc(proxyHandler.ProxyRecommendation)
	protectedRouter.PathPrefix("/popular").HandlerFunc(proxyHandler.ProxyRecommendation)

	// Event Ingestion Service routes
	protectedRouter.PathPrefix("/events").HandlerFunc(proxyHandler.ProxyEventIngestion)

	return r
}

func New(
	config *di.Config,
	logger *zap.Logger,
	jwtHandler *libsdi.JWTHandler,
	returns *libsdi.ReturnManager,
) *App {
	userDBClient := clients.NewUserDatabaseClient(config.UserDatabaseServiceURL)
	recommendClient := clients.NewRecommendationClient(config.RecommendationServiceURL)
	eventIngestionClient := clients.NewEventIngestionClient(config.EventIngestionServiceURL)

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
