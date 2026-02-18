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

func newLikesHandler(db *mockDB) *handlers.LikesHandler {
	cfg := testConfig()
	return handlers.NewLikesHandler(zap.NewNop(), cfg, testReturns(cfg), db)
}

// ── GetLikesForMusic ──────────────────────────────────────────────────────────

func TestGetLikesForMusic_Success(t *testing.T) {
	db := &mockDB{
		getLikesForMusicFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.PublicUser, error) {
			return []sqlhandler.PublicUser{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/music/"+testMusicUUID+"/likes", nil), map[string]string{"uuid": testMusicUUID})
	newLikesHandler(db).GetLikesForMusic(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestGetLikesForMusic_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/music/bad/likes", nil), map[string]string{"uuid": "bad"})
	newLikesHandler(&mockDB{}).GetLikesForMusic(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── GetLikesForUser ───────────────────────────────────────────────────────────

func TestGetLikesForUser_Success(t *testing.T) {
	db := &mockDB{
		getLikesForUserFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.Music, error) {
			return []sqlhandler.Music{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/"+testUserUUID+"/likes", nil), map[string]string{"uuid": testUserUUID})
	newLikesHandler(db).GetLikesForUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── IsLiked ───────────────────────────────────────────────────────────────────

func TestIsLiked_True(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		isLikedFn: func(_ context.Context, _ sqlhandler.IsLikedParams) (bool, error) {
			return true, nil
		},
	}
	h := handlers.NewLikesHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/music/"+testMusicUUID+"/liked", nil), map[string]string{"uuid": testMusicUUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.IsLiked(w, r)
	assertStatus(t, w, http.StatusOK)
	assertJSONBool(t, w, "liked", true)
}

func TestIsLiked_False(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		isLikedFn: func(_ context.Context, _ sqlhandler.IsLikedParams) (bool, error) {
			return false, nil
		},
	}
	h := handlers.NewLikesHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/music/"+testMusicUUID+"/liked", nil), map[string]string{"uuid": testMusicUUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.IsLiked(w, r)
	assertStatus(t, w, http.StatusOK)
	assertJSONBool(t, w, "liked", false)
}

// ── LikeMusic ─────────────────────────────────────────────────────────────────

func TestLikeMusic_Success(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewLikesHandler(zap.NewNop(), cfg, testReturns(cfg), &mockDB{})
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodPost, "/music/"+testMusicUUID+"/like", nil), map[string]string{"uuid": testMusicUUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.LikeMusic(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestLikeMusic_DBError(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		likeMusicFn: func(_ context.Context, _ sqlhandler.LikeMusicParams) error { return errDB },
	}
	h := handlers.NewLikesHandler(zap.NewNop(), cfg, testReturns(cfg), db)
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodPost, "/music/"+testMusicUUID+"/like", nil), map[string]string{"uuid": testMusicUUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.LikeMusic(w, r)
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── UnlikeMusic ───────────────────────────────────────────────────────────────

func TestUnlikeMusic_Success(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewLikesHandler(zap.NewNop(), cfg, testReturns(cfg), &mockDB{})
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodDelete, "/music/"+testMusicUUID+"/like", nil), map[string]string{"uuid": testMusicUUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UnlikeMusic(w, r)
	assertStatus(t, w, http.StatusOK)
}
