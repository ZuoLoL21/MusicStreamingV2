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

func newAlbumHandler(db *mockDB) *handlers.AlbumHandler {
	cfg := testConfig()
	return handlers.NewAlbumHandler(zap.NewNop(), cfg, testReturns(cfg), db)
}

// ── GetAlbum ──────────────────────────────────────────────────────────────────

func TestGetAlbum_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/albums/bad", nil), map[string]string{"uuid": "bad"})
	newAlbumHandler(&mockDB{}).GetAlbum(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetAlbum_NotFound(t *testing.T) {
	db := &mockDB{
		getAlbumFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Album, error) {
			return sqlhandler.Album{}, errors.New("not found")
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/albums/"+testAlbumUUID, nil), map[string]string{"uuid": testAlbumUUID})
	newAlbumHandler(db).GetAlbum(w, r)
	assertStatus(t, w, http.StatusNotFound)
}

func TestGetAlbum_Success(t *testing.T) {
	db := &mockDB{
		getAlbumFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Album, error) {
			return sqlhandler.Album{OriginalName: "Greatest Hits"}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/albums/"+testAlbumUUID, nil), map[string]string{"uuid": testAlbumUUID})
	newAlbumHandler(db).GetAlbum(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── GetAlbumsForArtist ────────────────────────────────────────────────────────

func TestGetAlbumsForArtist_Success(t *testing.T) {
	db := &mockDB{
		getAlbumsForArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.Album, error) {
			return []sqlhandler.Album{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/artists/"+testArtistUUID+"/albums", nil), map[string]string{"uuid": testArtistUUID})
	newAlbumHandler(db).GetAlbumsForArtist(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── CreateAlbum ───────────────────────────────────────────────────────────────

func TestCreateAlbum_Forbidden(t *testing.T) {
	// User is not a member of the artist
	cfg := testConfig()
	db := &mockDB{
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return []sqlhandler.GetUsersRepresentingArtistRow{}, nil
		},
	}
	h := handlers.NewAlbumHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPut, "/albums", map[string]string{
		"artist_uuid":   testArtistUUID,
		"original_name": "New Album",
	})
	r = withUserUUID(r, cfg, testUserUUID)
	h.CreateAlbum(w, r)
	assertStatus(t, w, http.StatusForbidden)
}

func TestCreateAlbum_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return ownerMembers(testUserUUID), nil
		},
	}
	h := handlers.NewAlbumHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPut, "/albums", map[string]string{
		"artist_uuid":   testArtistUUID,
		"original_name": "New Album",
	})
	r = withUserUUID(r, cfg, testUserUUID)
	h.CreateAlbum(w, r)
	assertStatus(t, w, http.StatusCreated)
}

// ── UpdateAlbum ───────────────────────────────────────────────────────────────

func TestUpdateAlbum_Forbidden(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getAlbumFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Album, error) {
			return sqlhandler.Album{FromArtist: mustUUID(testArtistUUID)}, nil
		},
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return []sqlhandler.GetUsersRepresentingArtistRow{}, nil
		},
	}
	h := handlers.NewAlbumHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/albums/"+testAlbumUUID, map[string]string{"original_name": "Updated"}),
		map[string]string{"uuid": testAlbumUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateAlbum(w, r)
	assertStatus(t, w, http.StatusForbidden)
}

func TestUpdateAlbum_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getAlbumFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Album, error) {
			return sqlhandler.Album{FromArtist: mustUUID(testArtistUUID)}, nil
		},
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return ownerMembers(testUserUUID), nil
		},
	}
	h := handlers.NewAlbumHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/albums/"+testAlbumUUID, map[string]string{"original_name": "Updated"}),
		map[string]string{"uuid": testAlbumUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateAlbum(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── DeleteAlbum ───────────────────────────────────────────────────────────────

func TestDeleteAlbum_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getAlbumFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Album, error) {
			return sqlhandler.Album{FromArtist: mustUUID(testArtistUUID)}, nil
		},
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return ownerMembers(testUserUUID), nil
		},
	}
	h := handlers.NewAlbumHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodDelete, "/albums/"+testAlbumUUID, nil), map[string]string{"uuid": testAlbumUUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.DeleteAlbum(w, r)
	assertStatus(t, w, http.StatusOK)
}
