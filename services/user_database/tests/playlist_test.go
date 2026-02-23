package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/handlers"
	sqlhandler "backend/sql/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func newPlaylistHandler(db *mockDB) *handlers.PlaylistHandler {
	cfg := testConfig()
	return handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), db)
}

// ── GetPlaylist ───────────────────────────────────────────────────────────────

func TestGetPlaylist_Success(t *testing.T) {
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{OriginalName: "My Playlist"}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/playlists/"+testPlaylistUUID, nil), map[string]string{"uuid": testPlaylistUUID})
	newPlaylistHandler(db).GetPlaylist(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestGetPlaylist_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/playlists/bad", nil), map[string]string{"uuid": "bad"})
	newPlaylistHandler(&mockDB{}).GetPlaylist(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetPlaylist_NotFound(t *testing.T) {
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{}, errDB
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/playlists/"+testPlaylistUUID, nil), map[string]string{"uuid": testPlaylistUUID})
	newPlaylistHandler(db).GetPlaylist(w, r)
	assertStatus(t, w, http.StatusNotFound)
}

// ── GetPlaylistsForUser ───────────────────────────────────────────────────────

func TestGetPlaylistsForUser_Success(t *testing.T) {
	db := &mockDB{
		getPlaylistsForUserFn: func(_ context.Context, _ sqlhandler.GetPlaylistsForUserParams) ([]sqlhandler.Playlist, error) {
			return []sqlhandler.Playlist{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/"+testUserUUID+"/playlists", nil), map[string]string{"uuid": testUserUUID})
	newPlaylistHandler(db).GetPlaylistsForUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── GetPlaylistTracks ─────────────────────────────────────────────────────────

func TestGetPlaylistTracks_Success(t *testing.T) {
	db := &mockDB{
		getPlaylistTracksFn: func(_ context.Context, _ sqlhandler.GetPlaylistTracksParams) ([]sqlhandler.Music, error) {
			return []sqlhandler.Music{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/playlists/"+testPlaylistUUID+"/tracks", nil), map[string]string{"uuid": testPlaylistUUID})
	newPlaylistHandler(db).GetPlaylistTracks(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── CreatePlaylist ────────────────────────────────────────────────────────────

func TestCreatePlaylist_Success(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), &mockDB{})
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPut, "/playlists", map[string]string{"original_name": "Chill Vibes"})
	r = withUserUUID(r, cfg, testUserUUID)
	h.CreatePlaylist(w, r)
	assertStatus(t, w, http.StatusCreated)
}

func TestCreatePlaylist_ValidationFail(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), &mockDB{})
	w := httptest.NewRecorder()
	// original_name is required but missing
	r := newRequest(http.MethodPut, "/playlists", map[string]string{})
	r = withUserUUID(r, cfg, testUserUUID)
	h.CreatePlaylist(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── UpdatePlaylist ────────────────────────────────────────────────────────────

func TestUpdatePlaylist_Forbidden(t *testing.T) {
	cfg := testConfig()
	// Playlist owned by a different user
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{FromUser: mustUUID(testUser2UUID)}, nil
		},
	}
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/playlists/"+testPlaylistUUID, map[string]string{"original_name": "Updated"}),
		map[string]string{"uuid": testPlaylistUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdatePlaylist(w, r)
	assertStatus(t, w, http.StatusForbidden)
}

func TestUpdatePlaylist_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{FromUser: mustUUID(testUserUUID)}, nil
		},
	}
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/playlists/"+testPlaylistUUID, map[string]string{"original_name": "Updated"}),
		map[string]string{"uuid": testPlaylistUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdatePlaylist(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── DeletePlaylist ────────────────────────────────────────────────────────────

func TestDeletePlaylist_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{FromUser: mustUUID(testUserUUID)}, nil
		},
	}
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodDelete, "/playlists/"+testPlaylistUUID, nil), map[string]string{"uuid": testPlaylistUUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.DeletePlaylist(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── AddTrackToPlaylist ────────────────────────────────────────────────────────

func TestAddTrackToPlaylist_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{FromUser: mustUUID(testUserUUID)}, nil
		},
	}
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPut, "/playlists/"+testPlaylistUUID+"/tracks/"+testMusicUUID, map[string]interface{}{"position": 0}),
		map[string]string{"uuid": testPlaylistUUID, "musicUuid": testMusicUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.AddTrackToPlaylist(w, r)
	assertStatus(t, w, http.StatusCreated)
}

func TestAddTrackToPlaylist_InvalidMusicUUID(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{FromUser: mustUUID(testUserUUID)}, nil
		},
	}
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPut, "/playlists/"+testPlaylistUUID+"/tracks/bad", map[string]interface{}{"position": 0}),
		map[string]string{"uuid": testPlaylistUUID, "musicUuid": "bad"},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.AddTrackToPlaylist(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── RemoveTrackFromPlaylist ───────────────────────────────────────────────────

func TestRemoveTrackFromPlaylist_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{FromUser: mustUUID(testUserUUID)}, nil
		},
	}
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodDelete, "/playlists/"+testPlaylistUUID+"/tracks/"+testMusicUUID, nil),
		map[string]string{"uuid": testPlaylistUUID, "musicUuid": testMusicUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.RemoveTrackFromPlaylist(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── UpdatePlaylistImage ───────────────────────────────────────────────────────

func TestUpdatePlaylistImage_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{FromUser: mustUUID(testUserUUID)}, nil
		},
	}
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/playlists/"+testPlaylistUUID+"/image", map[string]string{"image_path": "/img/cover.png"}),
		map[string]string{"uuid": testPlaylistUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdatePlaylistImage(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestUpdatePlaylistImage_Forbidden(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{FromUser: mustUUID(testUser2UUID)}, nil
		},
	}
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/playlists/"+testPlaylistUUID+"/image", map[string]string{"image_path": "/img/cover.png"}),
		map[string]string{"uuid": testPlaylistUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdatePlaylistImage(w, r)
	assertStatus(t, w, http.StatusForbidden)
}

func TestUpdatePlaylistImage_ValidationFail(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{FromUser: mustUUID(testUserUUID)}, nil
		},
	}
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/playlists/"+testPlaylistUUID+"/image", map[string]string{}),
		map[string]string{"uuid": testPlaylistUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdatePlaylistImage(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── UpdateTrackPosition ───────────────────────────────────────────────────────

func TestUpdateTrackPosition_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getPlaylistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Playlist, error) {
			return sqlhandler.Playlist{FromUser: mustUUID(testUserUUID)}, nil
		},
	}
	h := handlers.NewPlaylistHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/playlists/"+testPlaylistUUID+"/tracks/"+testTrackUUID+"/position", map[string]interface{}{"position": 2}),
		map[string]string{"uuid": testPlaylistUUID, "trackUuid": testTrackUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateTrackPosition(w, r)
	assertStatus(t, w, http.StatusOK)
}
