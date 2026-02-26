package handlers

import (
	"backend/internal/di"
	"backend/internal/storage"
	sqlhandler "backend/sql/sqlc"
	"net/http"
	"strings"

	libsdi "libs/di"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type SearchHandler struct {
	logger      *zap.Logger
	config      *di.Config
	returns     *libsdi.ReturnManager
	db          DB
	fileStorage storage.FileStorageClient
}

func NewSearchHandler(logger *zap.Logger, config *di.Config, returns *libsdi.ReturnManager, db DB, fileStorage storage.FileStorageClient) *SearchHandler {
	return &SearchHandler{
		logger:      logger,
		config:      config,
		returns:     returns,
		db:          db,
		fileStorage: fileStorage,
	}
}

type UserSearchResult struct {
	Uuid             string  `json:"uuid"`
	Username         string  `json:"username"`
	Email            string  `json:"email"`
	Bio              *string `json:"bio,omitempty"`
	ProfileImagePath *string `json:"profile_image_path,omitempty"`
	SimilarityScore  float64 `json:"similarity_score"`
}
type ArtistSearchResult struct {
	Uuid             string  `json:"uuid"`
	ArtistName       string  `json:"artist_name"`
	Bio              *string `json:"bio,omitempty"`
	ProfileImagePath *string `json:"profile_image_path,omitempty"`
	SimilarityScore  float64 `json:"similarity_score"`
}
type AlbumSearchResult struct {
	Uuid            string  `json:"uuid"`
	FromArtist      string  `json:"from_artist"`
	OriginalName    string  `json:"original_name"`
	Description     *string `json:"description,omitempty"`
	ImagePath       *string `json:"image_path,omitempty"`
	SimilarityScore float64 `json:"similarity_score"`
}
type MusicSearchResult struct {
	Uuid              string  `json:"uuid"`
	FromArtist        string  `json:"from_artist"`
	UploadedBy        string  `json:"uploaded_by"`
	InAlbum           *string `json:"in_album,omitempty"`
	SongName          string  `json:"song_name"`
	PathInFileStorage string  `json:"path_in_file_storage"`
	ImagePath         *string `json:"image_path,omitempty"`
	PlayCount         int32   `json:"play_count"`
	DurationSeconds   int32   `json:"duration_seconds"`
	SimilarityScore   float64 `json:"similarity_score"`
}
type PlaylistSearchResult struct {
	Uuid            string  `json:"uuid"`
	FromUser        string  `json:"from_user"`
	OriginalName    string  `json:"original_name"`
	Description     *string `json:"description,omitempty"`
	IsPublic        bool    `json:"is_public"`
	ImagePath       *string `json:"image_path,omitempty"`
	SimilarityScore float64 `json:"similarity_score"`
}

// SearchUsers searches for users by username or email
func (h *SearchHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" || len(query) < 2 {
		h.returns.ReturnError(w, "search query required (min 2 chars)", http.StatusBadRequest)
		return
	}

	limit, cursorScore, cursorTS := parsePaginationSearch(r)

	ctx := r.Context()
	users, err := h.db.SearchForUser(ctx, sqlhandler.SearchForUserParams{
		Similarity: query,
		Limit:      limit,
		Column3:    cursorScore.Float64,
		CreatedAt:  pgtype.Timestamp{Time: cursorTS.Time, Valid: cursorTS.Valid},
	})
	if err != nil {
		h.logger.Error("failed to search users", zap.Error(err))
		h.returns.ReturnError(w, "failed to search users", http.StatusInternalServerError)
		return
	}

	result := struct {
		Users []UserSearchResult `json:"users"`
	}{
		Users: h.formatUserResults(users),
	}

	h.returns.ReturnJSON(w, result, http.StatusOK)
}

// SearchArtists searches for artists by name
func (h *SearchHandler) SearchArtists(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" || len(query) < 2 {
		h.returns.ReturnError(w, "search query required (min 2 chars)", http.StatusBadRequest)
		return
	}

	limit, cursorScore, cursorTS := parsePaginationSearch(r)

	ctx := r.Context()
	artists, err := h.db.SearchForArtist(ctx, sqlhandler.SearchForArtistParams{
		Similarity: query,
		Limit:      limit,
		Column3:    cursorScore.Float64,
		CreatedAt:  pgtype.Timestamp{Time: cursorTS.Time, Valid: cursorTS.Valid},
	})
	if err != nil {
		h.logger.Error("failed to search artists", zap.Error(err))
		h.returns.ReturnError(w, "failed to search artists", http.StatusInternalServerError)
		return
	}

	result := struct {
		Artists []ArtistSearchResult `json:"artists"`
	}{
		Artists: h.formatArtistResults(artists),
	}

	h.returns.ReturnJSON(w, result, http.StatusOK)
}

// SearchAlbums searches for albums by name
func (h *SearchHandler) SearchAlbums(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" || len(query) < 2 {
		h.returns.ReturnError(w, "search query required (min 2 chars)", http.StatusBadRequest)
		return
	}

	limit, cursorScore, cursorTS := parsePaginationSearch(r)

	ctx := r.Context()
	albums, err := h.db.SearchForAlbum(ctx, sqlhandler.SearchForAlbumParams{
		Similarity: query,
		Limit:      limit,
		Column3:    cursorScore.Float64,
		CreatedAt:  pgtype.Timestamp{Time: cursorTS.Time, Valid: cursorTS.Valid},
	})
	if err != nil {
		h.logger.Error("failed to search albums", zap.Error(err))
		h.returns.ReturnError(w, "failed to search albums", http.StatusInternalServerError)
		return
	}

	result := struct {
		Albums []AlbumSearchResult `json:"albums"`
	}{
		Albums: h.formatAlbumResults(albums),
	}

	h.returns.ReturnJSON(w, result, http.StatusOK)
}

// SearchMusic searches for music tracks by name
func (h *SearchHandler) SearchMusic(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" || len(query) < 2 {
		h.returns.ReturnError(w, "search query required (min 2 chars)", http.StatusBadRequest)
		return
	}

	limit, cursorScore, cursorTS := parsePaginationSearch(r)

	ctx := r.Context()
	music, err := h.db.SearchForMusic(ctx, sqlhandler.SearchForMusicParams{
		Similarity: query,
		Limit:      limit,
		Column3:    cursorScore.Float64,
		CreatedAt:  pgtype.Timestamp{Time: cursorTS.Time, Valid: cursorTS.Valid},
	})
	if err != nil {
		h.logger.Error("failed to search music", zap.Error(err))
		h.returns.ReturnError(w, "failed to search music", http.StatusInternalServerError)
		return
	}

	result := struct {
		Music []MusicSearchResult `json:"music"`
	}{
		Music: h.formatMusicResults(music),
	}

	h.returns.ReturnJSON(w, result, http.StatusOK)
}

// SearchPlaylists searches for playlists by name
func (h *SearchHandler) SearchPlaylists(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := userUUIDFromCtx(w, r, h.config, h.returns)
	if !ok {
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" || len(query) < 2 {
		h.returns.ReturnError(w, "search query required (min 2 chars)", http.StatusBadRequest)
		return
	}

	limit, cursorScore, cursorTS := parsePaginationSearch(r)

	ctx := r.Context()
	playlists, err := h.db.SearchForPlaylist(ctx, sqlhandler.SearchForPlaylistParams{
		Similarity: query,
		Limit:      limit,
		FromUser:   userUUID,
		Column4:    cursorScore.Float64,
		CreatedAt:  pgtype.Timestamp{Time: cursorTS.Time, Valid: cursorTS.Valid},
	})
	if err != nil {
		h.logger.Error("failed to search playlists", zap.Error(err))
		h.returns.ReturnError(w, "failed to search playlists", http.StatusInternalServerError)
		return
	}

	result := struct {
		Playlists []PlaylistSearchResult `json:"playlists"`
	}{
		Playlists: h.formatPlaylistResults(playlists),
	}

	h.returns.ReturnJSON(w, result, http.StatusOK)
}

func (h *SearchHandler) formatUserResults(users []sqlhandler.SearchForUserRow) []UserSearchResult {
	results := make([]UserSearchResult, len(users))
	for i, u := range users {
		results[i] = UserSearchResult{
			Uuid:             uuidToString(u.Uuid),
			Username:         u.Username,
			Email:            u.Email,
			Bio:              pgtypeTextToPtr(u.Bio),
			ProfileImagePath: h.formatImagePath(u.ProfileImagePath, "user"),
			SimilarityScore:  parseSimilarityScore(u.SimilarityScore),
		}
	}
	return results
}

func (h *SearchHandler) formatArtistResults(artists []sqlhandler.SearchForArtistRow) []ArtistSearchResult {
	results := make([]ArtistSearchResult, len(artists))
	for i, a := range artists {
		results[i] = ArtistSearchResult{
			Uuid:             uuidToString(a.Uuid),
			ArtistName:       a.ArtistName,
			Bio:              pgtypeTextToPtr(a.Bio),
			ProfileImagePath: h.formatImagePath(a.ProfileImagePath, "artist"),
			SimilarityScore:  parseSimilarityScore(a.SimilarityScore),
		}
	}
	return results
}

func (h *SearchHandler) formatAlbumResults(albums []sqlhandler.SearchForAlbumRow) []AlbumSearchResult {
	results := make([]AlbumSearchResult, len(albums))
	for i, a := range albums {
		results[i] = AlbumSearchResult{
			Uuid:            uuidToString(a.Uuid),
			FromArtist:      uuidToString(a.FromArtist),
			OriginalName:    a.OriginalName,
			Description:     pgtypeTextToPtr(a.Description),
			ImagePath:       h.formatImagePath(a.ImagePath, "album"),
			SimilarityScore: parseSimilarityScore(a.SimilarityScore),
		}
	}
	return results
}

func (h *SearchHandler) formatMusicResults(music []sqlhandler.SearchForMusicRow) []MusicSearchResult {
	results := make([]MusicSearchResult, len(music))
	for i, m := range music {
		var inAlbum *string
		if m.InAlbum.Valid {
			s := uuidToString(m.InAlbum)
			inAlbum = &s
		}

		results[i] = MusicSearchResult{
			Uuid:              uuidToString(m.Uuid),
			FromArtist:        uuidToString(m.FromArtist),
			UploadedBy:        uuidToString(m.UploadedBy),
			InAlbum:           inAlbum,
			SongName:          m.SongName,
			PathInFileStorage: m.PathInFileStorage,
			ImagePath:         h.formatImagePath(m.ImagePath, "music"),
			PlayCount:         m.PlayCount.Int32,
			DurationSeconds:   m.DurationSeconds,
			SimilarityScore:   parseSimilarityScore(m.SimilarityScore),
		}
	}
	return results
}

func (h *SearchHandler) formatPlaylistResults(playlists []sqlhandler.SearchForPlaylistRow) []PlaylistSearchResult {
	results := make([]PlaylistSearchResult, len(playlists))
	for i, p := range playlists {
		results[i] = PlaylistSearchResult{
			Uuid:            uuidToString(p.Uuid),
			FromUser:        uuidToString(p.FromUser),
			OriginalName:    p.OriginalName,
			Description:     pgtypeTextToPtr(p.Description),
			IsPublic:        p.IsPublic.Bool,
			ImagePath:       h.formatImagePath(p.ImagePath, "playlist"),
			SimilarityScore: parseSimilarityScore(p.SimilarityScore),
		}
	}
	return results
}

func (h *SearchHandler) formatImagePath(path pgtype.Text, entityType string) *string {
	if !path.Valid || path.String == "" {
		defaultURL := h.fileStorage.GetDefaultImageURL(entityType)
		return &defaultURL
	}
	return &path.String
}

func uuidToString(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}
	return pgtype.UUID{Bytes: uuid.Bytes, Valid: true}.String()
}

func pgtypeTextToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func parseSimilarityScore(score interface{}) float64 {
	switch v := score.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	case int:
		return float64(v)
	default:
		return 0
	}
}
