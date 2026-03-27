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
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	popularityHandler, err := handlers.NewPopularityHandler(a.config, a.returns)
	if err != nil {
		a.logger.Fatal("failed to initialize popularity handler", zap.Error(err))
	}
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
		libsmiddleware.MetricsMiddleware(a.logger),
		libsmiddleware.FailureRecoveryMiddleware(a.logger),
		libsmiddleware.Logger(a.logger),
	)
	protectedRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.MetricsMiddleware(a.logger),
		libsmiddleware.FailureRecoveryMiddleware(a.logger),
		serviceAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.logger),
	)

	// Public routes
	r.Handle("/metrics", metrics.Handler()).Methods("GET")
	publicRouter.HandleFunc("/health", libshandlers.NewHealthCheckHandler("service-popularity-system")).Methods("GET")

	// All-time popularity endpoints
	protectedRouter.HandleFunc("/popular/songs/all-time", popularityHandler.PopularSongsAllTime).Methods("GET")
	protectedRouter.HandleFunc("/popular/artists/all-time", popularityHandler.PopularArtistAllTime).Methods("GET")
	protectedRouter.HandleFunc("/popular/themes/all-time", popularityHandler.PopularThemeAllTime).Methods("GET")
	protectedRouter.HandleFunc("/popular/songs/theme/{theme}", popularityHandler.PopularSongsAllTimeByTheme).Methods("GET")

	// Timeframe popularity endpoints
	protectedRouter.HandleFunc("/popular/songs/timeframe", popularityHandler.PopularSongsTimeframe).Methods("GET")
	protectedRouter.HandleFunc("/popular/artists/timeframe", popularityHandler.PopularArtistTimeframe).Methods("GET")
	protectedRouter.HandleFunc("/popular/themes/timeframe", popularityHandler.PopularThemeTimeframe).Methods("GET")
	protectedRouter.HandleFunc("/popular/songs/theme/{theme}/timeframe", popularityHandler.PopularSongsTimeframeByTheme).Methods("GET")

	return r
}

func New(logger *zap.Logger, config *di.Config, jwtHandler *libsdi.JWTHandler, returns *libsdi.ReturnManager) *App {
	return &App{
		logger:     logger,
		config:     config,
		jwtHandler: jwtHandler,
		returns:    returns,
	}
}
