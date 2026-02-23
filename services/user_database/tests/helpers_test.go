package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"backend/internal/di"

	libsdi "libs/di"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// ── well-known UUIDs used across all test files ───────────────────────────────

const (
	testUserUUID     = "11111111-1111-1111-1111-111111111111"
	testArtistUUID   = "22222222-2222-2222-2222-222222222222"
	testAlbumUUID    = "33333333-3333-3333-3333-333333333333"
	testMusicUUID    = "44444444-4444-4444-4444-444444444444"
	testPlaylistUUID = "55555555-5555-5555-5555-555555555555"
	testTrackUUID    = "66666666-6666-6666-6666-666666666666"
	testUser2UUID    = "77777777-7777-7777-7777-777777777777"
)

// errDB is a generic DB error returned by mocks in failure scenarios.
var errDB = errors.New("db error")

// ── config / dependency helpers ───────────────────────────────────────────────

// testConfig returns a *di.Config ready for unit tests.
// di.LoadConfig only reads env vars — no network calls happen.
func testConfig() *di.Config {
	_ = os.Setenv("TIME_IN_M_NORMAL", "15")
	_ = os.Setenv("TIME_IN_D_REFRESH", "7")
	_ = os.Setenv("SUBJECT_NORMAL", "access")
	_ = os.Setenv("SUBJECT_REFRESH", "refresh")
	return di.LoadConfig(zap.NewNop())
}

// testReturns creates a ReturnManager backed by a nop logger.
func testReturns(cfg *di.Config) *libsdi.ReturnManager {
	return libsdi.NewReturnManager(zap.NewNop())
}

// ── request helpers ───────────────────────────────────────────────────────────

// withUserUUID injects a user UUID string into the request context exactly as
// the auth middleware does at runtime.
func withUserUUID(r *http.Request, cfg *di.Config, uuidStr string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), cfg.UserUUIDKey, uuidStr))
}

// newRequest builds an httptest.Request with an optional JSON body.
func newRequest(method, target string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	return httptest.NewRequest(method, target, &buf)
}

// withVars injects gorilla/mux route variables so handlers calling mux.Vars(r)
// receive the expected values.
func withVars(r *http.Request, vars map[string]string) *http.Request {
	return mux.SetURLVars(r, vars)
}

// mustUUID converts a string to pgtype.UUID or panics.
func mustUUID(s string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		panic(err)
	}
	return id
}

// ── assertion helpers ─────────────────────────────────────────────────────────

// assertStatus fails the test if the recorder status differs from want.
func assertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status: want %d, got %d — body: %s", want, w.Code, w.Body.String())
	}
}

// assertJSONBool decodes the response as a JSON object and checks that key
// holds the given bool value.
func assertJSONBool(t *testing.T, w *httptest.ResponseRecorder, key string, want bool) {
	t.Helper()
	var m map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("decode JSON: %v — body: %s", err, w.Body.String())
	}
	got, ok := m[key]
	if !ok {
		t.Fatalf("key %q not in response %v", key, m)
	}
	if b, _ := got.(bool); b != want {
		t.Errorf("field %q: want %v, got %v", key, want, b)
	}
}
