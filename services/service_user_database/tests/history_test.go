package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/handlers"
	sqlhandler "backend/sql/sqlc"

	"go.uber.org/zap"
)

// ── GetListeningHistoryForUser ────────────────────────────────────────────────

func TestGetListeningHistoryForUser_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getListeningHistoryForUserFn: func(_ context.Context, _ sqlhandler.GetListeningHistoryForUserParams) ([]sqlhandler.ListeningHistory, error) {
			return []sqlhandler.ListeningHistory{}, nil
		},
	}
	h := handlers.NewHistoryHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodGet, "/me/history", nil)
	r = withUserUUID(r, cfg, testUserUUID)
	h.GetListeningHistoryForUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestGetListeningHistoryForUser_DBError(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getListeningHistoryForUserFn: func(_ context.Context, _ sqlhandler.GetListeningHistoryForUserParams) ([]sqlhandler.ListeningHistory, error) {
			return nil, errDB
		},
	}
	h := handlers.NewHistoryHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodGet, "/me/history", nil)
	r = withUserUUID(r, cfg, testUserUUID)
	h.GetListeningHistoryForUser(w, r)
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── GetTopMusicForUser ────────────────────────────────────────────────────────

func TestGetTopMusicForUser_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getTopMusicForUserFn: func(_ context.Context, _ sqlhandler.GetTopMusicForUserParams) ([]sqlhandler.GetTopMusicForUserRow, error) {
			return []sqlhandler.GetTopMusicForUserRow{}, nil
		},
	}
	h := handlers.NewHistoryHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodGet, "/me/history/top", nil)
	r = withUserUUID(r, cfg, testUserUUID)
	h.GetTopMusicForUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestGetTopMusicForUser_WithLimit(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewHistoryHandler(zap.NewNop(), cfg, testReturns(), &mockDB{})
	w := httptest.NewRecorder()
	r := newRequest(http.MethodGet, "/me/history/top?limit=20", nil)
	r = withUserUUID(r, cfg, testUserUUID)
	h.GetTopMusicForUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestGetTopMusicForUser_DBError(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getTopMusicForUserFn: func(_ context.Context, _ sqlhandler.GetTopMusicForUserParams) ([]sqlhandler.GetTopMusicForUserRow, error) {
			return nil, errDB
		},
	}
	h := handlers.NewHistoryHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodGet, "/me/history/top", nil)
	r = withUserUUID(r, cfg, testUserUUID)
	h.GetTopMusicForUser(w, r)
	assertStatus(t, w, http.StatusInternalServerError)
}
