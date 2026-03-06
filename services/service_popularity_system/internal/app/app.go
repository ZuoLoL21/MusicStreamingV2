package app

import (
	"libs/consts"
	"popularity/internal/di"
	"popularity/internal/handlers"

	libsdi "libs/di"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	Logger     *zap.Logger
	Config     *di.Config
	JWTHandler *libsdi.JWTHandler
	Returns    *libsdi.ReturnManager
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	popularityHandler := handlers.NewPopularityHandler(a.Logger, a.Config, a.Returns)
	serviceAuthHandler := libsmiddleware.NewAuthHandler(
		a.Logger,
		a.JWTHandler,
		a.Returns,
		consts.JWTSubjectService,
	)

	publicRouter := r.PathPrefix("").Subrouter()
	protectedRouter := r.PathPrefix("").Subrouter()

	publicRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.LoggingMiddleware(a.Logger),
		libsmiddleware.Logger(a.Logger),
	)
	protectedRouter.Use(
		libsmiddleware.RequestIDMiddleware(),
		libsmiddleware.LoggingMiddleware(a.Logger),
		serviceAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.Logger),
	)

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
		Logger:     logger,
		Config:     config,
		JWTHandler: jwtHandler,
		Returns:    returns,
	}
}
