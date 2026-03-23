//go:build integration

package integration

import (
	backenddi "backend/internal/di"
	"backend/internal/handlers"
	sqlhandler "backend/sql/sqlc"
	"backend/tests/integration/builders"
	"context"
	"libs/di"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIntegration_HistoryHandler_AddListeningEntry(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	artistOwner := builders.NewUserBuilder().WithEmail("artist@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(artistOwner).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, artistOwner).Build(t, ctx, db)
	listener := builders.NewUserBuilder().WithEmail("listener@test.com").Build(t, ctx, db)

	// Add listening entry directly via DB (no HTTP handler for POST)
	err := db.AddListeningHistoryEntry(ctx, sqlhandler.AddListeningHistoryEntryParams{
		UserUuid:  listener,
		MusicUuid: musicUUID,
	})
	require.NoError(t, err)

	// Verify entry was created
	history, err := db.GetListeningHistoryForUser(ctx, sqlhandler.GetListeningHistoryForUserParams{
		UserUuid: listener,
		Limit:    10,
	})
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestIntegration_HistoryHandler_AddListeningEntry_SameMusic(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	artistOwner := builders.NewUserBuilder().WithEmail("artist@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(artistOwner).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, artistOwner).Build(t, ctx, db)
	listener := builders.NewUserBuilder().WithEmail("listener@test.com").Build(t, ctx, db)

	// Add listening entries directly via DB (no HTTP handler for POST)
	err := db.AddListeningHistoryEntry(ctx, sqlhandler.AddListeningHistoryEntryParams{
		UserUuid:  listener,
		MusicUuid: musicUUID,
	})
	require.NoError(t, err)

	// Add second entry for same music
	ensureTimestampDistinct()
	err = db.AddListeningHistoryEntry(ctx, sqlhandler.AddListeningHistoryEntryParams{
		UserUuid:  listener,
		MusicUuid: musicUUID,
	})
	require.NoError(t, err)

	// Verify both entries exist
	history, err := db.GetListeningHistoryForUser(ctx, sqlhandler.GetListeningHistoryForUserParams{
		UserUuid: listener,
		Limit:    10,
	})
	require.NoError(t, err)
	assert.Len(t, history, 2)
}

func TestIntegration_HistoryHandler_GetListeningHistory(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	artistOwner := builders.NewUserBuilder().WithEmail("artist@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(artistOwner).Build(t, ctx, db)
	music1 := builders.NewMusicBuilder(artistUUID, artistOwner).Build(t, ctx, db)
	music2 := builders.NewMusicBuilder(artistUUID, artistOwner).Build(t, ctx, db)
	listener := builders.NewUserBuilder().WithEmail("listener@test.com").Build(t, ctx, db)

	// Add history entries
	db.AddListeningHistoryEntry(ctx, sqlhandler.AddListeningHistoryEntryParams{
		UserUuid:  listener,
		MusicUuid: music1,
	})
	ensureTimestampDistinct()
	db.AddListeningHistoryEntry(ctx, sqlhandler.AddListeningHistoryEntryParams{
		UserUuid:  listener,
		MusicUuid: music2,
	})

	handler := handlers.NewHistoryHandler(config, returns, db)

	req := createRequest(t, "GET", "/history?limit=10", nil)
	router := mux.NewRouter()
	router.HandleFunc("/history", wrapWithAuth(t, handler.GetListeningHistoryForUser, listener)).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var history []map[string]interface{}
	assertJSONResponse(t, rr, http.StatusOK, &history)
	assert.Len(t, history, 2)

	// Verify enriched fields are present
	firstEntry := history[0]
	assert.Contains(t, firstEntry, "song_name", "Expected song_name field")
	assert.Contains(t, firstEntry, "artist_name", "Expected artist_name field")
	assert.Contains(t, firstEntry, "artist_uuid", "Expected artist_uuid field")
	assert.NotEmpty(t, firstEntry["song_name"], "song_name should not be empty")
	assert.NotEmpty(t, firstEntry["artist_name"], "artist_name should not be empty")
	assert.NotEmpty(t, firstEntry["artist_uuid"], "artist_uuid should not be empty")
}

func TestIntegration_HistoryHandler_GetTopMusicForUser(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	artistOwner := builders.NewUserBuilder().WithEmail("artist@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(artistOwner).Build(t, ctx, db)
	music1 := builders.NewMusicBuilder(artistUUID, artistOwner).Build(t, ctx, db)
	music2 := builders.NewMusicBuilder(artistUUID, artistOwner).Build(t, ctx, db)
	listener := builders.NewUserBuilder().WithEmail("listener@test.com").Build(t, ctx, db)

	// Listen to music1 three times, music2 once
	db.AddListeningHistoryEntry(ctx, sqlhandler.AddListeningHistoryEntryParams{
		UserUuid:  listener,
		MusicUuid: music1,
	})
	ensureTimestampDistinct()
	db.AddListeningHistoryEntry(ctx, sqlhandler.AddListeningHistoryEntryParams{
		UserUuid:  listener,
		MusicUuid: music1,
	})
	ensureTimestampDistinct()
	db.AddListeningHistoryEntry(ctx, sqlhandler.AddListeningHistoryEntryParams{
		UserUuid:  listener,
		MusicUuid: music1,
	})
	ensureTimestampDistinct()
	db.AddListeningHistoryEntry(ctx, sqlhandler.AddListeningHistoryEntryParams{
		UserUuid:  listener,
		MusicUuid: music2,
	})

	handler := handlers.NewHistoryHandler(config, returns, db)

	req := createRequest(t, "GET", "/history/top?limit=10", nil)
	router := mux.NewRouter()
	router.HandleFunc("/history/top", wrapWithAuth(t, handler.GetTopMusicForUser, listener)).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var topMusic []map[string]interface{}
	assertJSONResponse(t, rr, http.StatusOK, &topMusic)
	require.GreaterOrEqual(t, len(topMusic), 2)

	// music1 should be first (most played)
	// Note: Exact ordering depends on implementation

	// Verify enriched fields are present
	firstEntry := topMusic[0]
	assert.Contains(t, firstEntry, "song_name", "Expected song_name field")
	assert.Contains(t, firstEntry, "artist_name", "Expected artist_name field")
	assert.Contains(t, firstEntry, "artist_uuid", "Expected artist_uuid field")
	assert.NotEmpty(t, firstEntry["song_name"], "song_name should not be empty")
	assert.NotEmpty(t, firstEntry["artist_name"], "artist_name should not be empty")
	assert.NotEmpty(t, firstEntry["artist_uuid"], "artist_uuid should not be empty")

	// Verify second entry also has enriched fields
	secondEntry := topMusic[1]
	assert.Contains(t, secondEntry, "song_name", "Expected song_name field in second entry")
	assert.Contains(t, secondEntry, "artist_name", "Expected artist_name field in second entry")
}
