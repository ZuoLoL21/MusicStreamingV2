package handlers

import (
	libsdi "libs/di"
	"net/http"
	"popularity/internal/di"

	"go.uber.org/zap"
)

type PopularityHandler struct {
	logger  *zap.Logger
	config  *di.Config
	returns *libsdi.ReturnManager
	db      *libsdi.DBHandler
}

func NewPopularityHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, db *libsdi.DBHandler) *PopularityHandler {
	return &PopularityHandler{
		logger:  logger,
		config:  config,
		returns: returns,
		db:      db,
	}
}

func (h *PopularityHandler) PopularSongsAllTime(w http.ResponseWriter, r *http.Request) {}

func (h *PopularityHandler) PopularSongsTimeframe(w http.ResponseWriter, r *http.Request) {}

func (h *PopularityHandler) PopularArtistAllTime(w http.ResponseWriter, r *http.Request) {}

func (h *PopularityHandler) PopularArtistTimeframe(w http.ResponseWriter, r *http.Request) {}

func (h *PopularityHandler) PopularThemeAllTime(w http.ResponseWriter, r *http.Request) {}

func (h *PopularityHandler) PopularThemeTimeframe(w http.ResponseWriter, r *http.Request) {}

func (h *PopularityHandler) PopularSongsAllTimeByTheme(w http.ResponseWriter, r *http.Request) {}

func (h *PopularityHandler) PopularSongsTimeframeByTheme(w http.ResponseWriter, r *http.Request) {}
