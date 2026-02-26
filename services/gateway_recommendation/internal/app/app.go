package app

import (
	"gateway_recommendation/internal/clients"
	"gateway_recommendation/internal/di"
	"gateway_recommendation/internal/handlers"

	libsdi "libs/di"
	libshandlers "libs/handlers"
	libshelpers "libs/helpers"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	config           *di.Config
	logger           *zap.Logger
	secrets          *libsdi.SecretsManager
	returnManager    *libsdi.ReturnManager
	banditClient     *clients.BanditClient
	popularityClient *clients.PopularityClient
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	// Create handlers
	recommendHandler := handlers.NewRecommendHandler(
		a.banditClient,
		a.popularityClient,
		a.returnManager,
		a.logger,
		a.config.RequestIDKey,
	)
	proxyHandler := handlers.NewProxyHandler(
		a.popularityClient,
		a.logger,
		a.config.RequestIDKey,
	)
	serviceAuthHandler := libsmiddleware.NewAuthHandler(
		a.logger,
		a.config,
		a.secrets,
		a.returnManager,
		libshelpers.JWTSubjectService,
	)

	protectedRouter := r.PathPrefix("").Subrouter()
	r.Use(
		libsmiddleware.RequestIDMiddleware(a.config),
		libsmiddleware.LoggingMiddleware(a.logger, a.config),
	)
	protectedRouter.Use(
		serviceAuthHandler.GetAuthMiddleware(),
	)

	// Health check route (no auth)
	r.HandleFunc("/", libshandlers.NewHealthCheckHandler("gateway-recommendation")).Methods("GET")

	// Recommendation endpoint
	protectedRouter.HandleFunc("/recommend/theme", recommendHandler.RecommendTheme).Methods("POST")

	// All-time popularity endpoints
	protectedRouter.HandleFunc("/popular/songs/all-time", proxyHandler.ProxyPopularSongsAllTime).Methods("GET")
	protectedRouter.HandleFunc("/popular/artists/all-time", proxyHandler.ProxyPopularArtistsAllTime).Methods("GET")
	protectedRouter.HandleFunc("/popular/themes/all-time", proxyHandler.ProxyPopularThemesAllTime).Methods("GET")
	protectedRouter.HandleFunc("/popular/songs/theme/{theme}", proxyHandler.ProxyPopularSongsByTheme).Methods("GET")

	// Timeframe popularity endpoints
	protectedRouter.HandleFunc("/popular/songs/timeframe", proxyHandler.ProxyPopularSongsTimeframe).Methods("GET")
	protectedRouter.HandleFunc("/popular/artists/timeframe", proxyHandler.ProxyPopularArtistsTimeframe).Methods("GET")
	protectedRouter.HandleFunc("/popular/themes/timeframe", proxyHandler.ProxyPopularThemesTimeframe).Methods("GET")
	protectedRouter.HandleFunc("/popular/songs/theme/{theme}/timeframe", proxyHandler.ProxyPopularSongsByThemeTimeframe).Methods("GET")

	return r
}

func NewApp(config *di.Config, logger *zap.Logger, secrets *libsdi.SecretsManager, returnManager *libsdi.ReturnManager) *App {
	banditClient := clients.NewBanditClient(config.BanditServiceURL, logger)
	popularityClient := clients.NewPopularityClient(config.PopularityServiceURL, logger)

	return &App{
		config:           config,
		logger:           logger,
		secrets:          secrets,
		returnManager:    returnManager,
		banditClient:     banditClient,
		popularityClient: popularityClient,
	}
}
