package app

import (
	"gateway_recommendation/internal/clients"
	"gateway_recommendation/internal/di"
	"gateway_recommendation/internal/handlers"

	libsdi "libs/di"
	libsmiddleware "libs/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	Config           *di.Config
	Logger           *zap.Logger
	ReturnManager    *libsdi.ReturnManager
	BanditClient     *clients.BanditClient
	PopularityClient *clients.PopularityClient
}

func (a *App) Router() *mux.Router {
	r := mux.NewRouter()

	recommendHandler := handlers.NewRecommendHandler(
		a.BanditClient,
		a.PopularityClient,
		a.ReturnManager,
		a.Logger,
		a.Config.RequestIDKey,
	)
	proxyHandler := handlers.NewProxyHandler(
		a.PopularityClient,
		a.Logger,
		a.Config.RequestIDKey,
	)

	r.Use(libsmiddleware.RequestIDMiddleware(a.Config))
	r.Use(libsmiddleware.LoggingMiddleware(a.Logger, a.Config))

	// Health check
	r.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// Recommendation endpoint
	r.HandleFunc("/recommend/theme", recommendHandler.RecommendTheme).Methods("POST")

	// All-time popularity endpoints
	r.HandleFunc("/popular/songs/all-time", proxyHandler.ProxyPopularSongsAllTime).Methods("GET")
	r.HandleFunc("/popular/artists/all-time", proxyHandler.ProxyPopularArtistsAllTime).Methods("GET")
	r.HandleFunc("/popular/themes/all-time", proxyHandler.ProxyPopularThemesAllTime).Methods("GET")
	r.HandleFunc("/popular/songs/theme/{theme}", proxyHandler.ProxyPopularSongsByTheme).Methods("GET")

	// Timeframe popularity endpoints
	r.HandleFunc("/popular/songs/timeframe", proxyHandler.ProxyPopularSongsTimeframe).Methods("GET")
	r.HandleFunc("/popular/artists/timeframe", proxyHandler.ProxyPopularArtistsTimeframe).Methods("GET")
	r.HandleFunc("/popular/themes/timeframe", proxyHandler.ProxyPopularThemesTimeframe).Methods("GET")
	r.HandleFunc("/popular/songs/theme/{theme}/timeframe", proxyHandler.ProxyPopularSongsByThemeTimeframe).Methods("GET")

	return r
}

func NewApp(config *di.Config, logger *zap.Logger, returnManager *libsdi.ReturnManager) *App {
	banditClient := clients.NewBanditClient(config.BanditServiceURL, logger)
	popularityClient := clients.NewPopularityClient(config.PopularityServiceURL, logger)

	return &App{
		Config:           config,
		Logger:           logger,
		ReturnManager:    returnManager,
		BanditClient:     banditClient,
		PopularityClient: popularityClient,
	}
}
