package app

import (
	"gateway_api/internal/clients"
	"gateway_api/internal/di"
	"gateway_api/internal/handlers"
	"gateway_api/internal/middleware"

	libsdi "libs/di"
	libshandlers "libs/handlers"
	libshelpers "libs/helpers"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	Config          *di.Config
	Logger          *zap.Logger
	Secrets         *libsdi.SecretsManager
	Returns         *libsdi.ReturnManager
	UserDBClient    *clients.UserDatabaseClient
	RecommendClient *clients.RecommendationClient
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware)

	// Create handlers
	proxyHandler := handlers.NewProxyHandler(
		a.UserDBClient,
		a.RecommendClient,
		a.Logger,
		a.Config.RequestIDKey,
		a.Config.ServiceJWTKey,
	)
	normalAuthHandler := libsmiddleware.NewAuthHandler(
		a.Logger,
		a.Config,
		a.Secrets,
		a.Returns,
		libshelpers.JWTSubjectNormal,
	)
	refreshAuthHandler := libsmiddleware.NewAuthHandler(
		a.Logger,
		a.Config,
		a.Secrets,
		a.Returns,
		libshelpers.JWTSubjectRefresh,
	)
	serviceJWTHandler := libsmiddleware.NewServiceJWTHandler(
		a.Logger,
		a.Config,
		a.Secrets,
		a.Returns,
		a.Config.JWTStorePath,
		a.Config.JWTExpirationService,
	)

	publicRouter := r.PathPrefix("").Subrouter()
	refreshRouter := r.PathPrefix("").Subrouter()
	protectedRouter := r.PathPrefix("").Subrouter()

	publicRouter.Use(
		libsmiddleware.RequestIDMiddleware(a.Config),
		libsmiddleware.LoggingMiddleware(a.Logger, a.Config),
		libsmiddleware.Logger(a.Logger, libsmiddleware.LoggerConfig{
			RequestIDKey: a.Config.RequestIDKey,
			UserUUIDKey:  a.Config.UserUUIDKey,
		}),
	)
	refreshRouter.Use(
		libsmiddleware.RequestIDMiddleware(a.Config),
		libsmiddleware.LoggingMiddleware(a.Logger, a.Config),
		refreshAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.Logger, libsmiddleware.LoggerConfig{
			RequestIDKey: a.Config.RequestIDKey,
			UserUUIDKey:  a.Config.UserUUIDKey,
		}),
		serviceJWTHandler.GetServiceJWTMiddleware(),
	)
	protectedRouter.Use(
		libsmiddleware.RequestIDMiddleware(a.Config),
		libsmiddleware.LoggingMiddleware(a.Logger, a.Config),
		normalAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.Logger, libsmiddleware.LoggerConfig{
			RequestIDKey: a.Config.RequestIDKey,
			UserUUIDKey:  a.Config.UserUUIDKey,
		}),
		serviceJWTHandler.GetServiceJWTMiddleware(),
	)

	// Public
	publicRouter.HandleFunc("/health", libshandlers.NewHealthCheckHandler("gateway-api")).Methods("GET")
	publicRouter.HandleFunc("/login", proxyHandler.ProxyLogin).Methods("POST", "PUT", "OPTIONS")

	// Renewal
	refreshRouter.HandleFunc("/renew", proxyHandler.ProxyRenew).Methods("POST")

	// Protected
	// User Database Service routes
	protectedRouter.PathPrefix("/users").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/artists").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/albums").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/music").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/tags").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/playlists").HandlerFunc(proxyHandler.ProxyUserDatabase)
	protectedRouter.PathPrefix("/history").HandlerFunc(proxyHandler.ProxyUserDatabase)

	// Recommendation Service routes
	protectedRouter.PathPrefix("/recommendation").HandlerFunc(proxyHandler.ProxyRecommendation)

	return r
}

func NewApp(
	config *di.Config,
	logger *zap.Logger,
	secrets *libsdi.SecretsManager,
	returns *libsdi.ReturnManager,
) *App {
	userDBClient := clients.NewUserDatabaseClient(config.UserDatabaseServiceURL, logger)
	recommendClient := clients.NewRecommendationClient(config.RecommendationServiceURL, logger)

	return &App{
		Config:          config,
		Logger:          logger,
		Secrets:         secrets,
		Returns:         returns,
		UserDBClient:    userDBClient,
		RecommendClient: recommendClient,
	}
}
