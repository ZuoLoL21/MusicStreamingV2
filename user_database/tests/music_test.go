package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/handlers"
	sqlhandler "backend/sql/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func newMusicHandler(db *mockDB) *handlers.MusicHandler {
	cfg := testConfig()
	return handlers.NewMusicHandler(zap.NewNop(), cfg, testReturns(cfg), db)
}

// ── GetMusic ──────────────────────────────────────────────────────────────────

func TestGetMusic_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/music/bad", nil), map[string]string{"uuid": "bad"})
	newMusicHandler(&mockDB{}).GetMusic(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetMusic_NotFound(t *testing.T) {
	db := &mockDB{
		getMusicFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Music, error) {
			return sqlhandler.Music{}, errors.New("not found")
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/music/"+testMusicUUID, nil), map[string]string{"uuid": testMusicUUID})
	newMusicHandler(db).GetMusic(w, r)
	assertStatus(t, w, http.StatusNotFound)
}

func TestGetMusic_Success(t *testing.T) {
	db := &mockDB{
		getMusicFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Music, error) {
			return sqlhandler.Music{SongName: "Song A"}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/music/"+testMusicUUID, nil), map[string]string{"uuid": testMusicUUID})
	newMusicHandler(db).GetMusic(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── GetMusicForArtist ─────────────────────────────────────────────────────────

func TestGetMusicForArtist_Success(t *testing.T) {
	db := &mockDB{
		getMusicForArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.Music, error) {
			return []sqlhandler.Music{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/artists/"+testArtistUUID+"/music", nil), map[string]string{"uuid": testArtistUUID})
	newMusicHandler(db).GetMusicForArtist(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── GetMusicForAlbum ──────────────────────────────────────────────────────────

func TestGetMusicForAlbum_Success(t *testing.T) {
	db := &mockDB{
		getMusicForAlbumFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.Music, error) {
			return []sqlhandler.Music{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/albums/"+testAlbumUUID+"/music", nil), map[string]string{"uuid": testAlbumUUID})
	newMusicHandler(db).GetMusicForAlbum(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── GetMusicForUser ───────────────────────────────────────────────────────────

func TestGetMusicForUser_Success(t *testing.T) {
	db := &mockDB{
		getMusicForUserFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.Music, error) {
			return []sqlhandler.Music{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/"+testUserUUID+"/music", nil), map[string]string{"uuid": testUserUUID})
	newMusicHandler(db).GetMusicForUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── CreateMusic ───────────────────────────────────────────────────────────────

func TestCreateMusic_Forbidden(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return []sqlhandler.GetUsersRepresentingArtistRow{}, nil
		},
	}
	h := handlers.NewMusicHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPut, "/music", map[string]interface{}{
		"artist_uuid":          testArtistUUID,
		"song_name":            "New Song",
		"path_in_file_storage": "/audio/new.mp3",
		"duration_seconds":     180,
	})
	r = withUserUUID(r, cfg, testUserUUID)
	h.CreateMusic(w, r)
	assertStatus(t, w, http.StatusForbidden)
}

func TestCreateMusic_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return ownerMembers(testUserUUID), nil
		},
	}
	h := handlers.NewMusicHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPut, "/music", map[string]interface{}{
		"artist_uuid":          testArtistUUID,
		"song_name":            "New Song",
		"path_in_file_storage": "/audio/new.mp3",
		"duration_seconds":     180,
	})
	r = withUserUUID(r, cfg, testUserUUID)
	h.CreateMusic(w, r)
	assertStatus(t, w, http.StatusCreated)
}

// ── UpdateMusicDetails ────────────────────────────────────────────────────────

func TestUpdateMusicDetails_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getMusicFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Music, error) {
			return sqlhandler.Music{FromArtist: mustUUID(testArtistUUID)}, nil
		},
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return ownerMembers(testUserUUID), nil
		},
	}
	h := handlers.NewMusicHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/music/"+testMusicUUID, map[string]string{"song_name": "Updated Song"}),
		map[string]string{"uuid": testMusicUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateMusicDetails(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── DeleteMusic ───────────────────────────────────────────────────────────────

func TestDeleteMusic_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getMusicFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Music, error) {
			return sqlhandler.Music{FromArtist: mustUUID(testArtistUUID)}, nil
		},
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return ownerMembers(testUserUUID), nil
		},
	}
	h := handlers.NewMusicHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodDelete, "/music/"+testMusicUUID, nil), map[string]string{"uuid": testMusicUUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.DeleteMusic(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── IncrementPlayCount ────────────────────────────────────────────────────────

func TestIncrementPlayCount_Success(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodPost, "/music/"+testMusicUUID+"/play", nil), map[string]string{"uuid": testMusicUUID})
	newMusicHandler(&mockDB{}).IncrementPlayCount(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestIncrementPlayCount_DBError(t *testing.T) {
	db := &mockDB{
		incrementPlayCountFn: func(_ context.Context, _ pgtype.UUID) error { return errDB },
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodPost, "/music/"+testMusicUUID+"/play", nil), map[string]string{"uuid": testMusicUUID})
	newMusicHandler(db).IncrementPlayCount(w, r)
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── AddListeningHistoryEntry ──────────────────────────────────────────────────

func TestAddListeningHistoryEntry_Success(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewMusicHandler(zap.NewNop(), cfg, testReturns(cfg), &mockDB{})
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/music/"+testMusicUUID+"/listen", map[string]interface{}{
			"listen_duration_seconds": 120,
			"completion_percentage":   0.8,
		}),
		map[string]string{"uuid": testMusicUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.AddListeningHistoryEntry(w, r)
	assertStatus(t, w, http.StatusCreated)
}
