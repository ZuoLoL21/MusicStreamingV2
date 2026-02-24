package handlers

import (
	"context"
	"database/sql"
	"fmt"
	libsdi "libs/di"
	"net/http"
	"popularity/internal/di"
	"strconv"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type PopularityHandler struct {
	logger      *zap.Logger
	config      *di.Config
	returns     *libsdi.ReturnManager
	warehouseDB *sql.DB
}

func NewPopularityHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager) *PopularityHandler {
	warehouseDB, err := sql.Open("clickhouse", config.WarehouseURL)
	if err != nil {
		logger.Fatal("failed to connect to ClickHouse", zap.Error(err))
	}

	return &PopularityHandler{
		logger:      logger,
		config:      config,
		returns:     returns,
		warehouseDB: warehouseDB,
	}
}

type SongPopularity struct {
	MusicUUID          string  `json:"music_uuid"`
	DecayPlays         float64 `json:"decay_plays"`
	DecayListenSeconds float64 `json:"decay_listen_seconds"`
}

type SongPopularityTimeframe struct {
	MusicUUID     string `json:"music_uuid"`
	Plays         uint64 `json:"plays"`
	ListenSeconds uint64 `json:"listen_seconds"`
}

type SongPopularityTheme struct {
	MusicUUID          string  `json:"music_uuid"`
	Theme              string  `json:"theme"`
	DecayPlays         float64 `json:"decay_plays"`
	DecayListenSeconds float64 `json:"decay_listen_seconds"`
}

type SongPopularityThemeTimeframe struct {
	MusicUUID     string `json:"music_uuid"`
	Theme         string `json:"theme"`
	Plays         uint64 `json:"plays"`
	ListenSeconds uint64 `json:"listen_seconds"`
}

type ArtistPopularity struct {
	ArtistUUID         string  `json:"artist_uuid"`
	DecayPlays         float64 `json:"decay_plays"`
	DecayListenSeconds float64 `json:"decay_listen_seconds"`
}

type ArtistPopularityTimeframe struct {
	ArtistUUID    string `json:"artist_uuid"`
	Plays         uint64 `json:"plays"`
	ListenSeconds uint64 `json:"listen_seconds"`
}

type ThemePopularity struct {
	Theme              string  `json:"theme"`
	DecayPlays         float64 `json:"decay_plays"`
	DecayListenSeconds float64 `json:"decay_listen_seconds"`
}

type ThemePopularityTimeframe struct {
	Theme         string `json:"theme"`
	Plays         uint64 `json:"plays"`
	ListenSeconds uint64 `json:"listen_seconds"`
}

// Scanner interface for row scanning
type Scanner interface {
	Scan(dest ...interface{}) error
}

// executeQuery executes a query and scans results using a provided scan function
func executeQuery[T any](
	ctx context.Context,
	db *sql.DB,
	logger *zap.Logger,
	query string,
	scanFn func(Scanner) (T, error),
	args ...interface{},
) ([]T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var results []T
	for rows.Next() {
		item, err := scanFn(rows)
		if err != nil {
			logger.Error("failed to scan row", zap.Error(err))
			continue
		}
		results = append(results, item)
	}

	return results, nil
}

func parseLimit(r *http.Request) int {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return 50 // default limit
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		return 50
	}
	return limit
}

func parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")

	if startStr == "" || endStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("start_date and end_date are required")
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date format")
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date format")
	}

	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date must be after start_date")
	}

	return start, end, nil
}

func (h *PopularityHandler) PopularSongsAllTime(w http.ResponseWriter, r *http.Request) {
	limit, cursorDecay, cursorID := parsePaginationDecay(r)

	query := `
		SELECT music_uuid, decay_plays, decay_listen_seconds
		FROM track_popularity
		WHERE (? = 0.0 OR (decay_plays < ? OR (decay_plays = ? AND music_uuid < ?)))
		ORDER BY decay_plays DESC, music_uuid DESC
		LIMIT ?
	`

	results, err := executeQuery(r.Context(), h.warehouseDB, h.logger, query,
		func(s Scanner) (SongPopularity, error) {
			var song SongPopularity
			err := s.Scan(&song.MusicUUID, &song.DecayPlays, &song.DecayListenSeconds)
			return song, err
		},
		cursorDecay, cursorDecay, cursorDecay, cursorID, limit,
	)

	if err != nil {
		h.logger.Error("failed to query popular songs", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular songs", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularSongsTimeframe(w http.ResponseWriter, r *http.Request) {
	limit, cursorPlays, cursorID := parsePaginationPlays(r)
	start, end, err := parseDateRange(r)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		SELECT music_uuid, sum(plays) AS total_plays, sum(listen_seconds) AS total_listen_seconds
		FROM track_popularity_daily
		WHERE for_day >= ? AND for_day <= ?
		GROUP BY music_uuid
		HAVING (? = 0 OR (total_plays < ? OR (total_plays = ? AND music_uuid < ?)))
		ORDER BY total_plays DESC, music_uuid DESC
		LIMIT ?
	`

	results, err := executeQuery(r.Context(), h.warehouseDB, h.logger, query,
		func(s Scanner) (SongPopularityTimeframe, error) {
			var song SongPopularityTimeframe
			err := s.Scan(&song.MusicUUID, &song.Plays, &song.ListenSeconds)
			return song, err
		},
		start, end, cursorPlays, cursorPlays, cursorPlays, cursorID, limit,
	)

	if err != nil {
		h.logger.Error("failed to query popular songs by timeframe", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular songs", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularArtistAllTime(w http.ResponseWriter, r *http.Request) {
	limit, cursorDecay, cursorID := parsePaginationDecay(r)

	query := `
		SELECT artist_uuid, decay_plays, decay_listen_seconds
		FROM artist_popularity
		WHERE (? = 0.0 OR (decay_plays < ? OR (decay_plays = ? AND artist_uuid < ?)))
		ORDER BY decay_plays DESC, artist_uuid DESC
		LIMIT ?
	`

	results, err := executeQuery(r.Context(), h.warehouseDB, h.logger, query,
		func(s Scanner) (ArtistPopularity, error) {
			var artist ArtistPopularity
			err := s.Scan(&artist.ArtistUUID, &artist.DecayPlays, &artist.DecayListenSeconds)
			return artist, err
		},
		cursorDecay, cursorDecay, cursorDecay, cursorID, limit,
	)

	if err != nil {
		h.logger.Error("failed to query popular artists", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular artists", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularArtistTimeframe(w http.ResponseWriter, r *http.Request) {
	limit, cursorPlays, cursorID := parsePaginationPlays(r)
	start, end, err := parseDateRange(r)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		SELECT artist_uuid, sum(plays) AS total_plays, sum(listen_seconds) AS total_listen_seconds
		FROM artist_popularity_daily
		WHERE for_day >= ? AND for_day <= ?
		GROUP BY artist_uuid
		HAVING (? = 0 OR (total_plays < ? OR (total_plays = ? AND artist_uuid < ?)))
		ORDER BY total_plays DESC, artist_uuid DESC
		LIMIT ?
	`

	results, err := executeQuery(r.Context(), h.warehouseDB, h.logger, query,
		func(s Scanner) (ArtistPopularityTimeframe, error) {
			var artist ArtistPopularityTimeframe
			err := s.Scan(&artist.ArtistUUID, &artist.Plays, &artist.ListenSeconds)
			return artist, err
		},
		start, end, cursorPlays, cursorPlays, cursorPlays, cursorID, limit,
	)

	if err != nil {
		h.logger.Error("failed to query popular artists by timeframe", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular artists", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularThemeAllTime(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r)

	query := `
		SELECT theme, decay_plays, decay_listen_seconds
		FROM theme_popularity
		ORDER BY decay_plays DESC
		LIMIT ?
	`

	results, err := executeQuery(r.Context(), h.warehouseDB, h.logger, query,
		func(s Scanner) (ThemePopularity, error) {
			var theme ThemePopularity
			err := s.Scan(&theme.Theme, &theme.DecayPlays, &theme.DecayListenSeconds)
			return theme, err
		},
		limit,
	)

	if err != nil {
		h.logger.Error("failed to query popular themes", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular themes", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularThemeTimeframe(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r)
	start, end, err := parseDateRange(r)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		SELECT theme, sum(plays) AS total_plays, sum(listen_seconds) AS total_listen_seconds
		FROM theme_popularity_daily
		WHERE for_day >= ? AND for_day <= ?
		GROUP BY theme
		ORDER BY total_plays DESC
		LIMIT ?
	`

	results, err := executeQuery(r.Context(), h.warehouseDB, h.logger, query,
		func(s Scanner) (ThemePopularityTimeframe, error) {
			var theme ThemePopularityTimeframe
			err := s.Scan(&theme.Theme, &theme.Plays, &theme.ListenSeconds)
			return theme, err
		},
		start, end, limit,
	)

	if err != nil {
		h.logger.Error("failed to query popular themes by timeframe", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular themes", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularSongsAllTimeByTheme(w http.ResponseWriter, r *http.Request) {
	limit, cursorDecay, cursorID := parsePaginationDecay(r)
	vars := mux.Vars(r)
	theme := vars["theme"]
	if theme == "" {
		h.returns.ReturnError(w, "theme parameter is required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT music_uuid, theme, decay_plays, decay_listen_seconds
		FROM track_by_theme_popularity
		WHERE theme = ? AND (? = 0.0 OR (decay_plays < ? OR (decay_plays = ? AND music_uuid < ?)))
		ORDER BY decay_plays DESC, music_uuid DESC
		LIMIT ?
	`

	results, err := executeQuery(r.Context(), h.warehouseDB, h.logger, query,
		func(s Scanner) (SongPopularityTheme, error) {
			var song SongPopularityTheme
			err := s.Scan(&song.MusicUUID, &song.Theme, &song.DecayPlays, &song.DecayListenSeconds)
			return song, err
		},
		theme, cursorDecay, cursorDecay, cursorDecay, cursorID, limit,
	)

	if err != nil {
		h.logger.Error("failed to query popular songs by theme", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular songs", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularSongsTimeframeByTheme(w http.ResponseWriter, r *http.Request) {
	limit, cursorPlays, cursorID := parsePaginationPlays(r)
	vars := mux.Vars(r)
	theme := vars["theme"]
	if theme == "" {
		h.returns.ReturnError(w, "theme parameter is required", http.StatusBadRequest)
		return
	}

	start, end, err := parseDateRange(r)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		SELECT music_uuid, theme, sum(plays) AS total_plays, sum(listen_seconds) AS total_listen_seconds
		FROM track_by_theme_popularity_daily
		WHERE theme = ? AND for_day >= ? AND for_day <= ?
		GROUP BY music_uuid, theme
		HAVING (? = 0 OR (total_plays < ? OR (total_plays = ? AND music_uuid < ?)))
		ORDER BY total_plays DESC, music_uuid DESC
		LIMIT ?
	`

	results, err := executeQuery(r.Context(), h.warehouseDB, h.logger, query,
		func(s Scanner) (SongPopularityThemeTimeframe, error) {
			var song SongPopularityThemeTimeframe
			err := s.Scan(&song.MusicUUID, &song.Theme, &song.Plays, &song.ListenSeconds)
			return song, err
		},
		theme, start, end, cursorPlays, cursorPlays, cursorPlays, cursorID, limit,
	)

	if err != nil {
		h.logger.Error("failed to query popular songs by theme and timeframe", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular songs", http.StatusInternalServerError)
		return
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}
