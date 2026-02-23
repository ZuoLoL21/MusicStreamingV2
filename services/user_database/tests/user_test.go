package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/handlers"
	sqlhandler "backend/sql/sqlc"

	libshelpers "libs/helpers"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func newUserHandler(db *mockDB) *handlers.UserHandler {
	cfg := testConfig()
	return handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), db)
}

// ── Register ──────────────────────────────────────────────────────────────────

func TestRegister_InvalidBody(t *testing.T) {
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPut, "/login", "not-json-at-all")
	newUserHandler(&mockDB{}).Register(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestRegister_ValidationFail(t *testing.T) {
	// username too short (min=5)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPut, "/login", map[string]string{
		"username": "ab",
		"email":    "user@example.com",
		"password": "password123",
	})
	newUserHandler(&mockDB{}).Register(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestRegister_DBError(t *testing.T) {
	db := &mockDB{
		createUserFn: func(_ context.Context, _ sqlhandler.CreateUserParams) (pgtype.UUID, error) {
			return pgtype.UUID{}, errors.New("unique violation")
		},
	}
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPut, "/login", map[string]string{
		"username": "username1",
		"email":    "user@example.com",
		"password": "password123",
	})
	newUserHandler(db).Register(w, r)
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── Login ─────────────────────────────────────────────────────────────────────

func TestLogin_InvalidBody(t *testing.T) {
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPost, "/login", "bad")
	newUserHandler(&mockDB{}).Login(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestLogin_UserNotFound(t *testing.T) {
	db := &mockDB{
		getUserByEmailFn: func(_ context.Context, _ string) (sqlhandler.User, error) {
			return sqlhandler.User{}, errors.New("not found")
		},
	}
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPost, "/login", map[string]string{
		"email":    "ghost@example.com",
		"password": "password123",
	})
	newUserHandler(db).Login(w, r)
	assertStatus(t, w, http.StatusUnauthorized)
}

// ── GetMe ─────────────────────────────────────────────────────────────────────

func TestGetMe_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getPublicUserFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.PublicUser, error) {
			return sqlhandler.PublicUser{Username: "alice", Email: "alice@example.com"}, nil
		},
	}
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), db)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodGet, "/users/me", nil)
	r = withUserUUID(r, cfg, testUserUUID)
	h.GetMe(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestGetMe_DBError(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getPublicUserFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.PublicUser, error) {
			return sqlhandler.PublicUser{}, errDB
		},
	}
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), db)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodGet, "/users/me", nil)
	r = withUserUUID(r, cfg, testUserUUID)
	h.GetMe(w, r)
	assertStatus(t, w, http.StatusNotFound)
}

// ── GetPublicUser ─────────────────────────────────────────────────────────────

func TestGetPublicUser_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/bad", nil), map[string]string{"uuid": "bad"})
	newUserHandler(&mockDB{}).GetPublicUser(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetPublicUser_NotFound(t *testing.T) {
	db := &mockDB{
		getPublicUserFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.PublicUser, error) {
			return sqlhandler.PublicUser{}, errors.New("not found")
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/"+testUserUUID, nil), map[string]string{"uuid": testUserUUID})
	newUserHandler(db).GetPublicUser(w, r)
	assertStatus(t, w, http.StatusNotFound)
}

func TestGetPublicUser_Success(t *testing.T) {
	db := &mockDB{
		getPublicUserFn: func(_ context.Context, _ pgtype.UUID) (sqlhandler.PublicUser, error) {
			return sqlhandler.PublicUser{Username: "alice"}, nil
		},
	}
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/"+testUserUUID, nil), map[string]string{"uuid": testUserUUID})
	newUserHandler(db).GetPublicUser(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── UpdateProfile ─────────────────────────────────────────────────────────────

func TestUpdateProfile_ValidationFail(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), &mockDB{})
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPost, "/users/me", map[string]string{"username": "ab"})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateProfile(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestUpdateProfile_Success(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), &mockDB{})
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPost, "/users/me", map[string]string{"username": "alice_updated"})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateProfile(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── UpdateEmail ───────────────────────────────────────────────────────────────

func TestUpdateEmail_InvalidEmail(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), &mockDB{})
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPost, "/users/me/email", map[string]string{"email": "not-an-email"})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateEmail(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestUpdateEmail_Success(t *testing.T) {
	password := "old_password"
	hashed := libshelpers.Encode(password)
	cfg := testConfig()
	df := &mockDB{
		getHashPasswordFn: func(_ context.Context, _ pgtype.UUID) (string, error) {
			return hashed, nil
		},
	}
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), df)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPost, "/users/me/email", map[string]string{
		"email":        "new@example.com",
		"old_password": password,
	})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateEmail(w, r)
	assertStatus(t, w, http.StatusOK)
}

func TestUpdateEmail_IncorrectPassword(t *testing.T) {
	cfg := testConfig()
	df := &mockDB{
		getHashPasswordFn: func(_ context.Context, _ pgtype.UUID) (string, error) {
			return libshelpers.Encode("random_password"), nil
		},
	}
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), df)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPost, "/users/me/email", map[string]string{
		"email":        "new@example.com",
		"old_password": "old_password",
	})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateEmail(w, r)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestUpdateEmail_MissingPassword(t *testing.T) {
	cfg := testConfig()
	df := &mockDB{
		getHashPasswordFn: func(_ context.Context, _ pgtype.UUID) (string, error) {
			return libshelpers.Encode("random_password"), nil
		},
	}
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), df)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPost, "/users/me/email", map[string]string{
		"email": "new@example.com",
	})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateEmail(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── UpdatePassword ────────────────────────────────────────────────────────────

func TestUpdatePassword_UserNotFound(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getHashPasswordFn: func(_ context.Context, _ pgtype.UUID) (string, error) {
			return "", errors.New("not found")
		},
	}
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), db)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPost, "/users/me/password", map[string]string{
		"old_password": "old12345678",
		"new_password": "new12345678",
	})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdatePassword(w, r)
	assertStatus(t, w, http.StatusNotFound)
}

func TestUpdatePassword_Success(t *testing.T) {
	cfg := testConfig()
	df := &mockDB{
		getHashPasswordFn: func(_ context.Context, _ pgtype.UUID) (string, error) {
			return libshelpers.Encode("old12345678"), nil
		},
	}
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), df)
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPost, "/users/me/password", map[string]string{
		"old_password": "old12345678",
		"new_password": "new12345678",
	})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdatePassword(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── UpdateImage ───────────────────────────────────────────────────────────────

func TestUpdateImage_Success(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), &mockDB{})
	w := httptest.NewRecorder()
	r := newRequest(http.MethodPost, "/users/me/image", map[string]string{"profile_image_path": "/img/avatar.png"})
	r = withUserUUID(r, cfg, testUserUUID)
	h.UpdateImage(w, r)
	assertStatus(t, w, http.StatusOK)
}

// ── GetArtistForUser ──────────────────────────────────────────────────────────

func TestGetArtistForUser_InvalidUUID(t *testing.T) {
	cfg := testConfig()
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), &mockDB{})
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/bad/artists", nil), map[string]string{"uuid": "bad"})
	h.GetArtistForUser(w, r)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetArtistForUser_Success(t *testing.T) {
	cfg := testConfig()
	db := &mockDB{
		getArtistForUserFn: func(_ context.Context, _ pgtype.UUID) ([]sqlhandler.GetArtistForUserRow, error) {
			return []sqlhandler.GetArtistForUserRow{}, nil
		},
	}
	h := handlers.NewUserHandler(zap.NewNop(), cfg, nil, testReturns(), db)
	w := httptest.NewRecorder()
	r := withVars(newRequest(http.MethodGet, "/users/"+testUserUUID+"/artists", nil), map[string]string{"uuid": testUserUUID})
	h.GetArtistForUser(w, r)
	assertStatus(t, w, http.StatusOK)
}
