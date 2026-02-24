package handlers

import (
	"database/sql"
	"fmt"
	libsdi "libs/di"
	"net/http"
	"popularity/internal/di"
	"strconv"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-chi/chi/v5"
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
	ctx := r.Context()
	limit, cursorDecay, cursorID := parsePaginationDecay(r)

	type SongPopularity struct {
		MusicUUID          string  `json:"music_uuid"`
		DecayPlays         float64 `json:"decay_plays"`
		DecayListenSeconds float64 `json:"decay_listen_seconds"`
	}

	query := `
		SELECT
			music_uuid,
			decay_plays,
			decay_listen_seconds
		FROM track_popularity
		WHERE (
			? = 0.0
			OR (
				decay_plays < ?
				OR (decay_plays = ? AND music_uuid < ?)
			)
		)
		ORDER BY decay_plays DESC, music_uuid DESC
		LIMIT ?
	`

	rows, err := h.warehouseDB.QueryContext(ctx, query, cursorDecay, cursorDecay, cursorDecay, cursorID, limit)
	if err != nil {
		h.logger.Error("failed to query popular songs", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular songs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := []SongPopularity{}
	for rows.Next() {
		var song SongPopularity
		if err := rows.Scan(&song.MusicUUID, &song.DecayPlays, &song.DecayListenSeconds); err != nil {
			h.logger.Error("failed to scan row", zap.Error(err))
			continue
		}
		results = append(results, song)
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularSongsTimeframe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, cursorPlays, cursorID := parsePaginationPlays(r)
	start, end, err := parseDateRange(r)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	type SongPopularity struct {
		MusicUUID     string `json:"music_uuid"`
		Plays         uint64 `json:"plays"`
		ListenSeconds uint64 `json:"listen_seconds"`
	}

	query := `
		SELECT
			music_uuid,
			sum(plays) AS total_plays,
			sum(listen_seconds) AS total_listen_seconds
		FROM track_popularity_daily
		WHERE for_day >= ? AND for_day <= ?
		GROUP BY music_uuid
		HAVING (
			? = 0
			OR (
				total_plays < ?
				OR (total_plays = ? AND music_uuid < ?)
			)
		)
		ORDER BY total_plays DESC, music_uuid DESC
		LIMIT ?
	`

	rows, err := h.warehouseDB.QueryContext(ctx, query, start, end, cursorPlays, cursorPlays, cursorPlays, cursorID, limit)
	if err != nil {
		h.logger.Error("failed to query popular songs by timeframe", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular songs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := []SongPopularity{}
	for rows.Next() {
		var song SongPopularity
		if err := rows.Scan(&song.MusicUUID, &song.Plays, &song.ListenSeconds); err != nil {
			h.logger.Error("failed to scan row", zap.Error(err))
			continue
		}
		results = append(results, song)
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularArtistAllTime(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, cursorDecay, cursorID := parsePaginationDecay(r)

	type ArtistPopularity struct {
		ArtistUUID         string  `json:"artist_uuid"`
		DecayPlays         float64 `json:"decay_plays"`
		DecayListenSeconds float64 `json:"decay_listen_seconds"`
	}

	query := `
		SELECT
			artist_uuid,
			decay_plays,
			decay_listen_seconds
		FROM artist_popularity
		WHERE (
			? = 0.0
			OR (
				decay_plays < ?
				OR (decay_plays = ? AND artist_uuid < ?)
			)
		)
		ORDER BY decay_plays DESC, artist_uuid DESC
		LIMIT ?
	`

	rows, err := h.warehouseDB.QueryContext(ctx, query, cursorDecay, cursorDecay, cursorDecay, cursorID, limit)
	if err != nil {
		h.logger.Error("failed to query popular artists", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular artists", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := []ArtistPopularity{}
	for rows.Next() {
		var artist ArtistPopularity
		if err := rows.Scan(&artist.ArtistUUID, &artist.DecayPlays, &artist.DecayListenSeconds); err != nil {
			h.logger.Error("failed to scan row", zap.Error(err))
			continue
		}
		results = append(results, artist)
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularArtistTimeframe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, cursorPlays, cursorID := parsePaginationPlays(r)
	start, end, err := parseDateRange(r)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	type ArtistPopularity struct {
		ArtistUUID    string `json:"artist_uuid"`
		Plays         uint64 `json:"plays"`
		ListenSeconds uint64 `json:"listen_seconds"`
	}

	query := `
		SELECT
			artist_uuid,
			sum(plays) AS total_plays,
			sum(listen_seconds) AS total_listen_seconds
		FROM artist_popularity_daily
		WHERE for_day >= ? AND for_day <= ?
		GROUP BY artist_uuid
		HAVING (
			? = 0
			OR (
				total_plays < ?
				OR (total_plays = ? AND artist_uuid < ?)
			)
		)
		ORDER BY total_plays DESC, artist_uuid DESC
		LIMIT ?
	`

	rows, err := h.warehouseDB.QueryContext(ctx, query, start, end, cursorPlays, cursorPlays, cursorPlays, cursorID, limit)
	if err != nil {
		h.logger.Error("failed to query popular artists by timeframe", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular artists", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := []ArtistPopularity{}
	for rows.Next() {
		var artist ArtistPopularity
		if err := rows.Scan(&artist.ArtistUUID, &artist.Plays, &artist.ListenSeconds); err != nil {
			h.logger.Error("failed to scan row", zap.Error(err))
			continue
		}
		results = append(results, artist)
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularThemeAllTime(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit := parseLimit(r)

	type ThemePopularity struct {
		Theme              string  `json:"theme"`
		DecayPlays         float64 `json:"decay_plays"`
		DecayListenSeconds float64 `json:"decay_listen_seconds"`
	}

	query := `
		SELECT
			theme,
			decay_plays,
			decay_listen_seconds
		FROM theme_popularity
		ORDER BY decay_plays DESC
		LIMIT ?
	`

	rows, err := h.warehouseDB.QueryContext(ctx, query, limit)
	if err != nil {
		h.logger.Error("failed to query popular themes", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular themes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := []ThemePopularity{}
	for rows.Next() {
		var theme ThemePopularity
		if err := rows.Scan(&theme.Theme, &theme.DecayPlays, &theme.DecayListenSeconds); err != nil {
			h.logger.Error("failed to scan row", zap.Error(err))
			continue
		}
		results = append(results, theme)
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularThemeTimeframe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit := parseLimit(r)
	start, end, err := parseDateRange(r)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	type ThemePopularity struct {
		Theme         string `json:"theme"`
		Plays         uint64 `json:"plays"`
		ListenSeconds uint64 `json:"listen_seconds"`
	}

	query := `
		SELECT
			theme,
			sum(plays) AS total_plays,
			sum(listen_seconds) AS total_listen_seconds
		FROM theme_popularity_daily
		WHERE for_day >= ? AND for_day <= ?
		GROUP BY theme
		ORDER BY total_plays DESC
		LIMIT ?
	`

	rows, err := h.warehouseDB.QueryContext(ctx, query, start, end, limit)
	if err != nil {
		h.logger.Error("failed to query popular themes by timeframe", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular themes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := []ThemePopularity{}
	for rows.Next() {
		var theme ThemePopularity
		if err := rows.Scan(&theme.Theme, &theme.Plays, &theme.ListenSeconds); err != nil {
			h.logger.Error("failed to scan row", zap.Error(err))
			continue
		}
		results = append(results, theme)
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularSongsAllTimeByTheme(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, cursorDecay, cursorID := parsePaginationDecay(r)
	theme := chi.URLParam(r, "theme")
	if theme == "" {
		h.returns.ReturnError(w, "theme parameter is required", http.StatusBadRequest)
		return
	}

	type SongPopularity struct {
		MusicUUID          string  `json:"music_uuid"`
		Theme              string  `json:"theme"`
		DecayPlays         float64 `json:"decay_plays"`
		DecayListenSeconds float64 `json:"decay_listen_seconds"`
	}

	query := `
		SELECT
			music_uuid,
			theme,
			decay_plays,
			decay_listen_seconds
		FROM track_by_theme_popularity
		WHERE theme = ?
			AND (
				? = 0.0
				OR (
					decay_plays < ?
					OR (decay_plays = ? AND music_uuid < ?)
				)
			)
		ORDER BY decay_plays DESC, music_uuid DESC
		LIMIT ?
	`

	rows, err := h.warehouseDB.QueryContext(ctx, query, theme, cursorDecay, cursorDecay, cursorDecay, cursorID, limit)
	if err != nil {
		h.logger.Error("failed to query popular songs by theme", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular songs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := []SongPopularity{}
	for rows.Next() {
		var song SongPopularity
		if err := rows.Scan(&song.MusicUUID, &song.Theme, &song.DecayPlays, &song.DecayListenSeconds); err != nil {
			h.logger.Error("failed to scan row", zap.Error(err))
			continue
		}
		results = append(results, song)
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}

func (h *PopularityHandler) PopularSongsTimeframeByTheme(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, cursorPlays, cursorID := parsePaginationPlays(r)
	theme := chi.URLParam(r, "theme")
	if theme == "" {
		h.returns.ReturnError(w, "theme parameter is required", http.StatusBadRequest)
		return
	}

	start, end, err := parseDateRange(r)
	if err != nil {
		h.returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return
	}

	type SongPopularity struct {
		MusicUUID     string `json:"music_uuid"`
		Theme         string `json:"theme"`
		Plays         uint64 `json:"plays"`
		ListenSeconds uint64 `json:"listen_seconds"`
	}

	query := `
		SELECT
			music_uuid,
			theme,
			sum(plays) AS total_plays,
			sum(listen_seconds) AS total_listen_seconds
		FROM track_by_theme_popularity_daily
		WHERE theme = ? AND for_day >= ? AND for_day <= ?
		GROUP BY music_uuid, theme
		HAVING (
			? = 0
			OR (
				total_plays < ?
				OR (total_plays = ? AND music_uuid < ?)
			)
		)
		ORDER BY total_plays DESC, music_uuid DESC
		LIMIT ?
	`

	rows, err := h.warehouseDB.QueryContext(ctx, query, theme, start, end, cursorPlays, cursorPlays, cursorPlays, cursorID, limit)
	if err != nil {
		h.logger.Error("failed to query popular songs by theme and timeframe", zap.Error(err))
		h.returns.ReturnError(w, "failed to fetch popular songs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := []SongPopularity{}
	for rows.Next() {
		var song SongPopularity
		if err := rows.Scan(&song.MusicUUID, &song.Theme, &song.Plays, &song.ListenSeconds); err != nil {
			h.logger.Error("failed to scan row", zap.Error(err))
			continue
		}
		results = append(results, song)
	}

	h.returns.ReturnJSON(w, results, http.StatusOK)
}
