package app

import (
	"popularity/internal/di"
	"popularity/internal/handlers"

	libsdi "libs/di"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	Logger  *zap.Logger
	Config  *di.Config
	Returns *libsdi.ReturnManager
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	popularityHandler := handlers.NewPopularityHandler(a.Logger, a.Config, a.Returns)

	r.Use(
		libsmiddleware.RequestIDMiddleware(a.Config),
		libsmiddleware.LoggingMiddleware(a.Logger, a.Config),
	)

	// All-time popularity endpoints
	r.HandleFunc("/popular/songs/all-time", popularityHandler.PopularSongsAllTime).Methods("GET")
	r.HandleFunc("/popular/artists/all-time", popularityHandler.PopularArtistAllTime).Methods("GET")
	r.HandleFunc("/popular/themes/all-time", popularityHandler.PopularThemeAllTime).Methods("GET")
	r.HandleFunc("/popular/songs/theme/{theme}", popularityHandler.PopularSongsAllTimeByTheme).Methods("GET")

	// Timeframe popularity endpoints
	r.HandleFunc("/popular/songs/timeframe", popularityHandler.PopularSongsTimeframe).Methods("GET")
	r.HandleFunc("/popular/artists/timeframe", popularityHandler.PopularArtistTimeframe).Methods("GET")
	r.HandleFunc("/popular/themes/timeframe", popularityHandler.PopularThemeTimeframe).Methods("GET")
	r.HandleFunc("/popular/songs/theme/{theme}/timeframe", popularityHandler.PopularSongsTimeframeByTheme).Methods("GET")

	return r
}

func New(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager) *App {
	return &App{Logger: logger, Config: config, Returns: returns}
}
