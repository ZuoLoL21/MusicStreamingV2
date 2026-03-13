package app

import (
	"gateway_recommendation/internal/clients"
	"gateway_recommendation/internal/di"
	"gateway_recommendation/internal/handlers"
	"libs/consts"

	libsdi "libs/di"
	libshandlers "libs/handlers"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	config           *di.Config
	logger           *zap.Logger
	jwtHandler       *libsdi.JWTHandler
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
	)
	proxyHandler := handlers.NewProxyHandler(
		a.popularityClient,
		a.logger,
	)
	serviceAuthHandler := libsmiddleware.NewAuthHandler(
		a.logger,
		a.jwtHandler,
		a.returnManager,
		consts.JWTSubjectService,
	)
	serviceJWTHandler := libsmiddleware.NewServiceJWTHandler(
		a.logger,
		a.jwtHandler,
		a.returnManager,
		a.config.JWTExpirationService,
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
		serviceJWTHandler.GetServiceJWTMiddleware(),
		libsmiddleware.Logger(a.logger),
	)

	// Health check route (no auth)
	publicRouter.HandleFunc("/", libshandlers.NewHealthCheckHandler("gateway-recommendation")).Methods("GET")

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

func NewApp(config *di.Config, logger *zap.Logger, jwtHandler *libsdi.JWTHandler, returnManager *libsdi.ReturnManager) *App {
	banditClient := clients.NewBanditClient(config.BanditServiceURL, logger)
	popularityClient := clients.NewPopularityClient(config.PopularityServiceURL, logger)

	return &App{
		config:           config,
		logger:           logger,
		jwtHandler:       jwtHandler,
		returnManager:    returnManager,
		banditClient:     banditClient,
		popularityClient: popularityClient,
	}
}
