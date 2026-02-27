package app

import (
	"popularity/internal/di"
	"popularity/internal/handlers"

	libsdi "libs/di"
	libshelpers "libs/helpers"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	Logger  *zap.Logger
	Config  *di.Config
	Secrets *libsdi.SecretsManager
	Returns *libsdi.ReturnManager
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	popularityHandler := handlers.NewPopularityHandler(a.Logger, a.Config, a.Returns)
	serviceAuthHandler := libsmiddleware.NewAuthHandler(
		a.Logger,
		a.Config,
		a.Secrets,
		a.Returns,
		libshelpers.JWTSubjectService,
	)

	publicRouter := r.PathPrefix("").Subrouter()
	protectedRouter := r.PathPrefix("").Subrouter()

	publicRouter.Use(
		libsmiddleware.RequestIDMiddleware(a.Config),
		libsmiddleware.LoggingMiddleware(a.Logger, a.Config),
		libsmiddleware.Logger(a.Logger, libsmiddleware.LoggerConfig{
			RequestIDKey: a.Config.RequestIDKey,
			UserUUIDKey:  a.Config.UserUUIDKey,
		}),
	)
	protectedRouter.Use(
		libsmiddleware.RequestIDMiddleware(a.Config),
		libsmiddleware.LoggingMiddleware(a.Logger, a.Config),
		serviceAuthHandler.GetAuthMiddleware(),
		libsmiddleware.Logger(a.Logger, libsmiddleware.LoggerConfig{
			RequestIDKey: a.Config.RequestIDKey,
			UserUUIDKey:  a.Config.UserUUIDKey,
		}),
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

func New(logger *zap.Logger, config *di.Config, secrets *libsdi.SecretsManager, returns *libsdi.ReturnManager) *App {
	return &App{
		Logger:  logger,
		Config:  config,
		Secrets: secrets,
		Returns: returns,
	}
}
