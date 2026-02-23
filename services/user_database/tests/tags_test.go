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

func newTagsHandler(db *mockDB) *handlers.TagsHandler {
	cfg := testConfig()
	return handlers.NewTagsHandler(zap.NewNop(), cfg, testReturns(), db)
}

// ── GetAllTags ────────────────────────────────────────────────────────────────

func TestGetAllTags_Success(t *testing.T) {
	db := &mockDB{
		getAllTagsFn: func(_ context.Context, _ sqlhandler.GetAllTagsParams) ([]sqlhandler.MusicTag, error) {
			return []sqlhandler.MusicTag{{TagName: "pop"}}, nil
		},
	}
	w := httptest.NewRecorder()
	newTagsHandler(db).GetAllTags(w, newRequest(http.MethodGet, "/tags", nil))
	assertStatus(t, w, http.StatusOK)
}

func TestGetAllTags_DBError(t *testing.T) {
	db := &mockDB{
		getAllTagsFn: func(_ context.Context, _ sqlhandler.GetAllTagsParams) ([]sqlhandler.MusicTag, error) {
			return nil, errDB
		},
	}
	w := httptest.NewRecorder()
	newTagsHandler(db).GetAllTags(w, newRequest(http.MethodGet, "/tags", nil))
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── GetTag ────────────────────────────────────────────────────────────────────

func TestGetTag_Success(t *testing.T) {
	db := &mockDB{
		getTagFn: func(_ context.Context, _ string) (sqlhandler.MusicTag, error) {
			return sqlhandler.MusicTag{TagName: "pop"}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/tags/pop", nil), map[string]string{"name": "pop"})
	newTagsHandler(db).GetTag(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestGetTag_NotFound(t *testing.T) {
	db := &mockDB{
		getTagFn: func(_ context.Context, _ string) (sqlhandler.MusicTag, error) {
			return sqlhandler.MusicTag{}, errDB
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/tags/unknown", nil), map[string]string{"name": "unknown"})
	newTagsHandler(db).GetTag(w, r)
	assertStatus(t, w, http.StatusNotFound)
}

// ── GetMusicForTag ────────────────────────────────────────────────────────────

func TestGetMusicForTag_Success(t *testing.T) {
	db := &mockDB{
		getMusicForTagFn: func(_ context.Context, _ sqlhandler.GetMusicForTagParams) ([]sqlhandler.Music, error) {
			return []sqlhandler.Music{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/tags/pop/music", nil), map[string]string{"name": "pop"})
	newTagsHandler(db).GetMusicForTag(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── GetTagsForMusic ───────────────────────────────────────────────────────────

func TestGetTagsForMusic_Success(t *testing.T) {
	db := &mockDB{
		getTagsForMusicFn: func(_ context.Context, _ sqlhandler.GetTagsForMusicParams) ([]sqlhandler.MusicTag, error) {
			return []sqlhandler.MusicTag{}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/music/"+testMusicUUID+"/tags", nil), map[string]string{"uuid": testMusicUUID})
	newTagsHandler(db).GetTagsForMusic(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestGetTagsForMusic_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/music/bad/tags", nil), map[string]string{"uuid": "bad"})
	newTagsHandler(&mockDB{}).GetTagsForMusic(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── CreateTag ─────────────────────────────────────────────────────────────────

func TestCreateTag_Success(t *testing.T) {
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPut, "/tags", map[string]string{"tag_name": "jazz"})
	newTagsHandler(&mockDB{}).CreateTag(w, r)
	assertStatus(t, w, http.StatusCreated)
}

func TestCreateTag_ValidationFail(t *testing.T) {
	w := httptest.NewRecorder()
	// tag_name is required but missing
	r := newRequest(http.MethodPut, "/tags", map[string]string{})
	newTagsHandler(&mockDB{}).CreateTag(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── AssignTagToMusic ──────────────────────────────────────────────────────────

func TestAssignTagToMusic_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getMusicFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Music, error) {
			return sqlhandler.Music{FromArtist: mustUUID(testArtistUUID)}, nil
		},
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return []sqlhandler.GetUsersRepresentingArtistRow{{Uuid: mustUUID(testUserUUID), Role: sqlhandler.ArtistMemberRoleMember}}, nil
		},
	}
	h := handlers.NewTagsHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodPost, "/music/"+testMusicUUID+"/tags/jazz", nil), map[string]string{"uuid": testMusicUUID, "name": "jazz"})
	r = withUserUUID(r, cfg, testUserUUID)
	h.AssignTagToMusic(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestAssignTagToMusic_Forbidden(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getMusicFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Music, error) {
			return sqlhandler.Music{FromArtist: mustUUID(testArtistUUID)}, nil
		},
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return []sqlhandler.GetUsersRepresentingArtistRow{}, nil
		},
	}
	h := handlers.NewTagsHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodPost, "/music/"+testMusicUUID+"/tags/jazz", nil), map[string]string{"uuid": testMusicUUID, "name": "jazz"})
	r = withUserUUID(r, cfg, testUserUUID)
	h.AssignTagToMusic(w, r)
	assertStatus(t, w, http.StatusForbidden)
}

func TestAssignTagToMusic_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodPost, "/music/bad/tags/jazz", nil), map[string]string{"uuid": "bad", "name": "jazz"})
	newTagsHandler(&mockDB{}).AssignTagToMusic(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── RemoveTagFromMusic ────────────────────────────────────────────────────────

func TestRemoveTagFromMusic_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getMusicFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Music, error) {
			return sqlhandler.Music{FromArtist: mustUUID(testArtistUUID)}, nil
		},
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return []sqlhandler.GetUsersRepresentingArtistRow{{Uuid: mustUUID(testUserUUID), Role: sqlhandler.ArtistMemberRoleMember}}, nil
		},
	}
	h := handlers.NewTagsHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodDelete, "/music/"+testMusicUUID+"/tags/jazz", nil), map[string]string{"uuid": testMusicUUID, "name": "jazz"})
	r = withUserUUID(r, cfg, testUserUUID)
	h.RemoveTagFromMusic(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestRemoveTagFromMusic_DBError(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getMusicFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.Music, error) {
			return sqlhandler.Music{FromArtist: mustUUID(testArtistUUID)}, nil
		},
		getUsersRepresentingArtistFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetUsersRepresentingArtistRow, error) {
			return []sqlhandler.GetUsersRepresentingArtistRow{{Uuid: mustUUID(testUserUUID), Role: sqlhandler.ArtistMemberRoleMember}}, nil
		},
		removeTagFromMusicFn: func(_ context.Context, _ sqlhandler.RemoveTagFromMusicParams) error {
			return errDB
		},
	}
	h := handlers.NewTagsHandler(zap.NewNop(), cfg, testReturns(), db)
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodDelete, "/music/"+testMusicUUID+"/tags/jazz", nil), map[string]string{"uuid": testMusicUUID, "name": "jazz"})
	r = withUserUUID(r, cfg, testUserUUID)
	h.RemoveTagFromMusic(w, r)
	assertStatus(t, w, http.StatusInternalServerError)
}
