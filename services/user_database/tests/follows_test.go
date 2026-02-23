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

func newFollowsHandler(db *mockDB) *handlers.FollowsHandler {
	cfg := testConfig()
	return handlers.NewFollowsHandler(zap.NewNop(), cfg, testReturns(cfg), db)
}

// ── GetFollowersForUser ───────────────────────────────────────────────────────

func TestGetFollowersForUser_Success(t *testing.T) {
	db := &mockDB{
		getFollowersForUserFn: func(_ context.Context, _ sqlhandler.GetFollowersForUserParams) ([]sqlhandler.PublicUser, error) {
			return []sqlhandler.PublicUser{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/"+testUserUUID+"/followers", nil), map[string]string{"uuid": testUserUUID})
	newFollowsHandler(db).GetFollowersForUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestGetFollowersForUser_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/bad/followers", nil), map[string]string{"uuid": "bad"})
	newFollowsHandler(&mockDB{}).GetFollowersForUser(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── GetFollowingUsersForUser ──────────────────────────────────────────────────

func TestGetFollowingUsersForUser_Success(t *testing.T) {
	db := &mockDB{
		getFollowedUsersForUserFn: func(_ context.Context, _ sqlhandler.GetFollowedUsersForUserParams) ([]sqlhandler.PublicUser, error) {
			return []sqlhandler.PublicUser{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/"+testUserUUID+"/following/users", nil), map[string]string{"uuid": testUserUUID})
	newFollowsHandler(db).GetFollowingUsersForUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestGetFollowingUsersForUser_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/bad/following/users", nil), map[string]string{"uuid": "bad"})
	newFollowsHandler(&mockDB{}).GetFollowingUsersForUser(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── GetFollowedArtistsForUser ─────────────────────────────────────────────────

func TestGetFollowedArtistsForUser_Success(t *testing.T) {
	db := &mockDB{
		getFollowedArtistsForUserFn: func(_ context.Context, _ sqlhandler.GetFollowedArtistsForUserParams) ([]sqlhandler.PublicUser, error) {
			return []sqlhandler.PublicUser{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/"+testUserUUID+"/following/artists", nil), map[string]string{"uuid": testUserUUID})
	newFollowsHandler(db).GetFollowedArtistsForUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestGetFollowedArtistsForUser_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/bad/following/artists", nil), map[string]string{"uuid": "bad"})
	newFollowsHandler(&mockDB{}).GetFollowedArtistsForUser(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── GetFollowersForArtist ─────────────────────────────────────────────────────

func TestGetFollowersForArtist_Success(t *testing.T) {
	db := &mockDB{
		getFollowersForArtistFn: func(_ context.Context, _ sqlhandler.GetFollowersForArtistParams) ([]sqlhandler.PublicUser, error) {
			return []sqlhandler.PublicUser{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/artists/"+testArtistUUID+"/followers", nil), map[string]string{"uuid": testArtistUUID})
	newFollowsHandler(db).GetFollowersForArtist(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── FollowUser ────────────────────────────────────────────────────────────────

func TestFollowUser_Success(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewFollowsHandler(zap.NewNop(), cfg, testReturns(cfg), &mockDB{})
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodPost, "/users/"+testUser2UUID+"/follow", nil), map[string]string{"uuid": testUser2UUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.FollowUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestFollowUser_InvalidUUID(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewFollowsHandler(zap.NewNop(), cfg, testReturns(cfg), &mockDB{})
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodPost, "/users/bad/follow", nil), map[string]string{"uuid": "bad"})
	r = withUserUUID(r, cfg, testUserUUID)
	h.FollowUser(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── UnfollowUser ──────────────────────────────────────────────────────────────

func TestUnfollowUser_Success(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewFollowsHandler(zap.NewNop(), cfg, testReturns(cfg), &mockDB{})
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodDelete, "/users/"+testUser2UUID+"/follow", nil), map[string]string{"uuid": testUser2UUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UnfollowUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── FollowArtist ──────────────────────────────────────────────────────────────

func TestFollowArtist_Success(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewFollowsHandler(zap.NewNop(), cfg, testReturns(cfg), &mockDB{})
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodPost, "/artists/"+testArtistUUID+"/follow", nil), map[string]string{"uuid": testArtistUUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.FollowArtist(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── UnfollowArtist ────────────────────────────────────────────────────────────

func TestUnfollowArtist_Success(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewFollowsHandler(zap.NewNop(), cfg, testReturns(cfg), &mockDB{})
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodDelete, "/artists/"+testArtistUUID+"/follow", nil), map[string]string{"uuid": testArtistUUID})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UnfollowArtist(w, r)
	assertStatus(t, w, http.StatusOK)
}
