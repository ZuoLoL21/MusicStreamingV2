package app

import (
	"gateway_recommendation/internal/clients"
	"gateway_recommendation/internal/di"
	"gateway_recommendation/internal/handlers"
	"libs/consts"
	"libs/metrics"

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
	handlers         *HandlerRegistry
}

type HandlerRegistry struct {
	Recommend *handlers.RecommendHandler
	Proxy     *handlers.ProxyHandler
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()
	a.initHandlers()

	normalRouter := r.PathPrefix("").Subrouter()
	protectedRouter := a.setupMiddleware(normalRouter)

	// Register routes
	a.registerMonitoringRoutes(r)
	a.registerRecommendationRoutes(protectedRouter)
	a.registerPopularityRoutes(protectedRouter)

	return r
}

func (a *App) initHandlers() {
	a.handlers = &HandlerRegistry{
		Recommend: handlers.NewRecommendHandler(
			a.banditClient,
			a.popularityClient,
			a.returnManager,
		),
		Proxy: handlers.NewProxyHandler(
			a.popularityClient,
		),
	}
}

func (a *App) setupMiddleware(normalRouter *mux.Router) *mux.Router {
	// Auth middleware setup
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
		serviceJWTHandler.GetServiceJWTMiddleware(),
	)

	return protectedRouter
}

func (a *App) registerMonitoringRoutes(r *mux.Router) {
	r.Handle("/metrics", metrics.Handler()).Methods("GET")
	r.HandleFunc("/health", libshandlers.NewHealthCheckHandler("gateway-recommendation")).Methods("GET")
}

func (a *App) registerRecommendationRoutes(r *mux.Router) {
	r.HandleFunc("/recommend/theme", a.handlers.Recommend.RecommendTheme).Methods("POST")
}

func (a *App) registerPopularityRoutes(r *mux.Router) {
	// All-time popularity endpoints
	r.HandleFunc("/popular/songs/all-time", a.handlers.Proxy.ProxyPopularSongsAllTime).Methods("GET")
	r.HandleFunc("/popular/artists/all-time", a.handlers.Proxy.ProxyPopularArtistsAllTime).Methods("GET")
	r.HandleFunc("/popular/themes/all-time", a.handlers.Proxy.ProxyPopularThemesAllTime).Methods("GET")
	r.HandleFunc("/popular/songs/theme/{theme}", a.handlers.Proxy.ProxyPopularSongsByTheme).Methods("GET")

	// Timeframe popularity endpoints
	r.HandleFunc("/popular/songs/timeframe", a.handlers.Proxy.ProxyPopularSongsTimeframe).Methods("GET")
	r.HandleFunc("/popular/artists/timeframe", a.handlers.Proxy.ProxyPopularArtistsTimeframe).Methods("GET")
	r.HandleFunc("/popular/themes/timeframe", a.handlers.Proxy.ProxyPopularThemesTimeframe).Methods("GET")
	r.HandleFunc("/popular/songs/theme/{theme}/timeframe", a.handlers.Proxy.ProxyPopularSongsByThemeTimeframe).Methods("GET")
}

func New(config *di.Config, logger *zap.Logger, jwtHandler *libsdi.JWTHandler, returnManager *libsdi.ReturnManager) *App {
	banditClient := clients.NewBanditClient(config.BanditServiceURL)
	popularityClient := clients.NewPopularityClient(config.PopularityServiceURL)

	return &App{
		config:           config,
		logger:           logger,
		jwtHandler:       jwtHandler,
		returnManager:    returnManager,
		banditClient:     banditClient,
		popularityClient: popularityClient,
	}
}
