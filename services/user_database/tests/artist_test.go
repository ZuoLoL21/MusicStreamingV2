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

func newArtistHandler(db *mockDB) *handlers.ArtistHandler {
	cfg := testConfig()
	return handlers.NewArtistHandler(zap.NewNop(), cfg, testReturns(), db)
}

// ownerMembers returns a GetUsersRepresentingArtistRow slice where userUUIDStr
// is an owner, used to satisfy checkArtistRole inside handlers.
func ownerMembers(userUUIDStr string) []sqlhandler.GetUsersRepresentingArtistRow {
	return []sqlhandler.GetUsersRepresentingArtistRow{
		{
			Uuid: mustUUID(userUUIDStr),
			Role: sqlhandler.ArtistMemberRoleOwner,
		},
	}
}

// ── GetArtistsAlphabetically ──────────────────────────────────────────────────

func TestGetArtistsAlphabetically_Success(t *testing.T) {
	db := &mockDB{
		getArtistsAlphabeticallyFn: func(_ context.Context, _ sqlhandler.GetArtistsAlphabeticallyParams) ([]sqlhandler.Artist, error) {
			return []sqlhandler.Artist{{ArtistName: "The Band"}}, nil
		},
	}
	w := httptest.NewRecorder()
	newArtistHandler(db).GetArtistsAlphabetically(w, newRequest(http.MethodGet, "/artists", nil))
	assertStatus(t, w, http.StatusOK)
}

func TestGetArtistsAlphabetically_DBError(t *testing.T) {
	db := &mockDB{
		getArtistsAlphabeticallyFn: func(_ context.Context, _ sqlhandler.GetArtistsAlphabeticallyParams) ([]sqlhandler.Artist, error) {
			return nil, errDB
		},
	}
	w := httptest.NewRecorder()
	newArtistHandler(db).GetArtistsAlphabetically(w, newRequest(http.MethodGet, "/artists", nil))
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── GetArtist ─────────────────────────────────────────────────────────────────

func TestGetArtist_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/artists/bad", nil), map[string]string{"uuid": "bad"})
	newArtistHandler(&mockDB{}).GetArtist(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetArtist_NotFound(t *testing.T) {
	db := &mockDB{
		getArtistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Artist, error) {
			return sqlhandler.Artist{}, errors.New("not found")
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/artists/"+testArtistUUID, nil), map[string]string{"uuid": testArtistUUID})
	newArtistHandler(db).GetArtist(w, r)
	assertStatus(t, w, http.StatusNotFound)
}

func TestGetArtist_Success(t *testing.T) {
	db := &mockDB{
		getArtistFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Artist, error) {
			return sqlhandler.Artist{ArtistName: "The Band"}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/artists/"+testArtistUUID, nil), map[string]string{"uuid": testArtistUUID})
	newArtistHandler(db).GetArtist(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── CreateArtist ──────────────────────────────────────────────────────────────

func TestCreateArtist_Success(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewArtistHandler(zap.NewNop(), cfg, testReturns(), &mockDB{})
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPut, "/artists", map[string]string{"artist_name": "The Band"})
	r = withUserUUID(r, cfg, testUserUUID)
	h.CreateArtist(w, r)
	assertStatus(t, w, http.StatusCreated)
}

func TestCreateArtist_ValidationFail(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewArtistHandler(zap.NewNop(), cfg, testReturns(), &mockDB{})
	w := httptest.NewRecorder()
	// artist_name missing (required)
	r := newRequest(http.MethodPut, "/artists", map[string]string{})
	r = withUserUUID(r, cfg, testUserUUID)
	h.CreateArtist(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── UpdateArtistProfile ───────────────────────────────────────────────────────

func TestUpdateArtistProfile_Forbidden(t *testing.T) {
	cfg := testConfig()
	// No members → checkArtistRole returns false
	db := &mockDB{
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return []sqlhandler.GetUsersRepresentingArtistRow{}, nil
		},
	}
	h := handlers.NewArtistHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/artists/"+testArtistUUID, map[string]string{"artist_name": "New Name"}),
		map[string]string{"uuid": testArtistUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateArtistProfile(w, r)
	assertStatus(t, w, http.StatusForbidden)
}

func TestUpdateArtistProfile_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return ownerMembers(testUserUUID), nil
		},
	}
	h := handlers.NewArtistHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/artists/"+testArtistUUID, map[string]string{"artist_name": "New Name"}),
		map[string]string{"uuid": testArtistUUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateArtistProfile(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── GetUsersRepresentingArtist ────────────────────────────────────────────────

func TestGetUsersRepresentingArtist_Success(t *testing.T) {
	db := &mockDB{
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return ownerMembers(testUserUUID), nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/artists/"+testArtistUUID+"/members", nil), map[string]string{"uuid": testArtistUUID})
	newArtistHandler(db).GetUsersRepresentingArtist(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── AddUserToArtist ───────────────────────────────────────────────────────────

func TestAddUserToArtist_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return ownerMembers(testUserUUID), nil
		},
	}
	h := handlers.NewArtistHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPut, "/artists/"+testArtistUUID+"/members/"+testUser2UUID, map[string]string{"role": "member"}),
		map[string]string{"uuid": testArtistUUID, "userUuid": testUser2UUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.AddUserToArtist(w, r)
	assertStatus(t, w, http.StatusCreated)
}

// ── RemoveUserFromArtist ──────────────────────────────────────────────────────

func TestRemoveUserFromArtist_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return ownerMembers(testUserUUID), nil
		},
	}
	h := handlers.NewArtistHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodDelete, "/artists/"+testArtistUUID+"/members/"+testUser2UUID, nil),
		map[string]string{"uuid": testArtistUUID, "userUuid": testUser2UUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.RemoveUserFromArtist(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── ChangeUserRole ────────────────────────────────────────────────────────────

func TestChangeUserRole_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return ownerMembers(testUserUUID), nil
		},
	}
	h := handlers.NewArtistHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := withVars(
		newRequest(http.MethodPost, "/artists/"+testArtistUUID+"/members/"+testUser2UUID+"/role", map[string]string{"role": "manager"}),
		map[string]string{"uuid": testArtistUUID, "userUuid": testUser2UUID},
	)
	r = withUserUUID(r, cfg, testUserUUID)
	h.ChangeUserRole(w, r)
	assertStatus(t, w, http.StatusOK)
}
