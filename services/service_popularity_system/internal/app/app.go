package app

import (
	"libs/consts"
	libsdi "libs/di"
	libshandlers "libs/handlers"
	"libs/metrics"
	libsmiddleware "libs/middleware"
	"popularity/internal/di"
	"popularity/internal/handlers"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	logger     *zap.Logger
	config     *di.Config
	jwtHandler *libsdi.JWTHandler
	returns    *libsdi.ReturnManager
	handlers   *HandlerRegistry
}

type HandlerRegistry struct {
	Popularity *handlers.PopularityHandler
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()
	a.initHandlers()

	normalRouter := r.PathPrefix("").Subrouter()
	protectedRouter := a.setupMiddleware(normalRouter)

	// Register routes
	a.registerMonitoringRoutes(r)
	a.registerPopularityRoutes(protectedRouter)

	return r
}

func (a *App) initHandlers() {
	popularityHandler, err := handlers.NewPopularityHandler(a.config, a.returns)
	if err != nil {
		a.logger.Fatal("failed to initialize popularity handler", zap.Error(err))
	}

	a.handlers = &HandlerRegistry{
		Popularity: popularityHandler,
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
	r.HandleFunc("/health", libshandlers.NewHealthCheckHandler("service-popularity-system")).Methods("GET")
}

func (a *App) registerPopularityRoutes(r *mux.Router) {
	// All-time popularity endpoints
	r.HandleFunc("/popular/songs/all-time", a.handlers.Popularity.PopularSongsAllTime).Methods("GET")
	r.HandleFunc("/popular/artists/all-time", a.handlers.Popularity.PopularArtistAllTime).Methods("GET")
	r.HandleFunc("/popular/themes/all-time", a.handlers.Popularity.PopularThemeAllTime).Methods("GET")
	r.HandleFunc("/popular/songs/theme/{theme}", a.handlers.Popularity.PopularSongsAllTimeByTheme).Methods("GET")

	// Timeframe popularity endpoints
	r.HandleFunc("/popular/songs/timeframe", a.handlers.Popularity.PopularSongsTimeframe).Methods("GET")
	r.HandleFunc("/popular/artists/timeframe", a.handlers.Popularity.PopularArtistTimeframe).Methods("GET")
	r.HandleFunc("/popular/themes/timeframe", a.handlers.Popularity.PopularThemeTimeframe).Methods("GET")
	r.HandleFunc("/popular/songs/theme/{theme}/timeframe", a.handlers.Popularity.PopularSongsTimeframeByTheme).Methods("GET")
}

func New(logger *zap.Logger, config *di.Config, jwtHandler *libsdi.JWTHandler, returns *libsdi.ReturnManager) *App {
	return &App{
		logger:     logger,
		config:     config,
		jwtHandler: jwtHandler,
		returns:    returns,
	}
}
