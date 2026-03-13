package app

import (
	"gateway_api/internal/clients"
	"gateway_api/internal/di"
	"gateway_api/internal/handlers"
	"libs/consts"

	libsdi "libs/di"
	libshandlers "libs/handlers"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	Config               *di.Config
	Logger               *zap.Logger
	JWTHandler           *libsdi.JWTHandler
	Returns              *libsdi.ReturnManager
	UserDBClient         *clients.UserDatabaseClient
	RecommendClient      *clients.RecommendationClient
	EventIngestionClient *clients.EventIngestionClient
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()
	r.Use(libsmiddleware.CORSMiddleware)

	// Create handlers
	proxyHandler := handlers.NewProxyHandler(
		a.UserDBClient,
		a.RecommendClient,
		a.EventIngestionClient,
		a.Logger,
	)
	normalAuthHandler := libsmiddleware.NewAuthHandler(
		a.Logger,
		a.JWTHandler,
		a.Returns,
		consts.JWTSubjectNormal,
	)
	refreshAuthHandler := libsmiddleware.NewAuthHandler(
		a.Logger,
		a.JWTHandler,
		a.Returns,
		consts.JWTSubjectRefresh,
	)
	serviceJWTHandler := libsmiddleware.NewServiceJWTHandler(
		a.Logger,
		a.JWTHandler,
		a.Returns,
		a.Config.JWTExpirationService,
	)

	publicRouter := r.PathPrefix("").Subrouter()
	refreshRouter := r.PathPrefix("").Subrouter()
	protectedRouter := r.PathPrefix("").Subrouter()

	publicRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.FailureRecoveryMiddleware(a.Logger),
		libsmiddleware.Logger(a.Logger),
	)
	refreshRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.FailureRecoveryMiddleware(a.Logger),
		refreshAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.Logger),
		serviceJWTHandler.GetServiceJWTMiddleware(),
	)
	protectedRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.FailureRecoveryMiddleware(a.Logger),
		normalAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.Logger),
		serviceJWTHandler.GetServiceJWTMiddleware(),
	)

	// Public
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

func NewApp(
	config *di.Config,
	logger *zap.Logger,
	jwtHandler *libsdi.JWTHandler,
	returns *libsdi.ReturnManager,
) *App {
	userDBClient := clients.NewUserDatabaseClient(config.UserDatabaseServiceURL, logger)
	recommendClient := clients.NewRecommendationClient(config.RecommendationServiceURL, logger)
	eventIngestionClient := clients.NewEventIngestionClient(config.EventIngestionServiceURL, logger)

	return &App{
		Config:               config,
		Logger:               logger,
		JWTHandler:           jwtHandler,
		Returns:              returns,
		UserDBClient:         userDBClient,
		RecommendClient:      recommendClient,
		EventIngestionClient: eventIngestionClient,
	}
}
