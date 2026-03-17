//go:build integration

package integration

import (
	backenddi "backend/internal/di"
	"backend/internal/handlers"
	sqlhandler "backend/sql/sqlc"
	"backend/tests/integration/builders"
	"context"
	"encoding/json"
	"libs/di"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIntegration_PlaylistHandler_ReorderTracks_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	music1UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music2UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music3UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	// Add tracks
	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music1UUID,
	})
	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music2UUID,
	})
	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music3UUID,
	})

	handler := handlers.NewPlaylistHandler(logger, config, returns, db, fileStorage)

	// Reorder: music3, music1, music2
	requestBody := map[string]interface{}{
		"track_order": []string{
			builders.UUIDToString(music3UUID),
			builders.UUIDToString(music1UUID),
			builders.UUIDToString(music2UUID),
		},
	}
	req := createJSONRequest(t, "POST", "/playlists/"+builders.UUIDToString(playlistUUID)+"/reorder", requestBody)

	router := mux.NewRouter()
	router.HandleFunc("/playlists/{uuid}/reorder", wrapWithAuth(t, handler.ReorderPlaylistTracks, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify order changed
	tracks, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        10,
		Column3:      -1,
	})
	require.NoError(t, err)
	require.Len(t, tracks, 3)
	assert.Equal(t, music3UUID, tracks[0].Uuid)
	assert.Equal(t, music1UUID, tracks[1].Uuid)
	assert.Equal(t, music2UUID, tracks[2].Uuid)
}

func TestIntegration_PlaylistHandler_ReorderTracks_EmptyArray(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	music1UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music1UUID,
	})

	handler := handlers.NewPlaylistHandler(logger, config, returns, db, fileStorage)

	// Try empty array
	requestBody := map[string]interface{}{
		"track_order": []string{},
	}
	req := createJSONRequest(t, "POST", "/playlists/"+builders.UUIDToString(playlistUUID)+"/reorder", requestBody)

	router := mux.NewRouter()
	router.HandleFunc("/playlists/{uuid}/reorder", wrapWithAuth(t, handler.ReorderPlaylistTracks, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "cannot be empty")
}

func TestIntegration_PlaylistHandler_ReorderTracks_DuplicateUUIDs(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	music1UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music2UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music3UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music1UUID,
	})
	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music2UUID,
	})
	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music3UUID,
	})

	handler := handlers.NewPlaylistHandler(logger, config, returns, db, fileStorage)

	// Try with duplicate UUID (music1 appears twice)
	music1Str := builders.UUIDToString(music1UUID)
	requestBody := map[string]interface{}{
		"track_order": []string{music1Str, music1Str, builders.UUIDToString(music2UUID)},
	}
	req := createJSONRequest(t, "POST", "/playlists/"+builders.UUIDToString(playlistUUID)+"/reorder", requestBody)

	router := mux.NewRouter()
	router.HandleFunc("/playlists/{uuid}/reorder", wrapWithAuth(t, handler.ReorderPlaylistTracks, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "duplicate")
}

func TestIntegration_PlaylistHandler_ReorderTracks_CountMismatch(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	music1UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music2UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music3UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music1UUID,
	})
	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music2UUID,
	})
	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music3UUID,
	})

	handler := handlers.NewPlaylistHandler(logger, config, returns, db, fileStorage)

	// Try with only 2 tracks (playlist has 3)
	requestBody := map[string]interface{}{
		"track_order": []string{
			builders.UUIDToString(music1UUID),
			builders.UUIDToString(music2UUID),
		},
	}
	req := createJSONRequest(t, "POST", "/playlists/"+builders.UUIDToString(playlistUUID)+"/reorder", requestBody)

	router := mux.NewRouter()
	router.HandleFunc("/playlists/{uuid}/reorder", wrapWithAuth(t, handler.ReorderPlaylistTracks, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "count")
}

func TestIntegration_PlaylistHandler_ReorderTracks_InvalidTrack(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	music1UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music2UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	musicNotInPlaylistUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music1UUID,
	})
	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music2UUID,
	})

	handler := handlers.NewPlaylistHandler(logger, config, returns, db, fileStorage)

	// Try with a track not in the playlist
	requestBody := map[string]interface{}{
		"track_order": []string{
			builders.UUIDToString(music1UUID),
			builders.UUIDToString(musicNotInPlaylistUUID),
		},
	}
	req := createJSONRequest(t, "POST", "/playlists/"+builders.UUIDToString(playlistUUID)+"/reorder", requestBody)

	router := mux.NewRouter()
	router.HandleFunc("/playlists/{uuid}/reorder", wrapWithAuth(t, handler.ReorderPlaylistTracks, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "not found")
}

func TestIntegration_PlaylistHandler_ReorderTracks_InvalidUUID(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	music1UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music1UUID,
	})

	handler := handlers.NewPlaylistHandler(logger, config, returns, db, fileStorage)

	// Try with invalid UUID format
	requestBody := map[string]interface{}{
		"track_order": []string{"not-a-valid-uuid"},
	}
	req := createJSONRequest(t, "POST", "/playlists/"+builders.UUIDToString(playlistUUID)+"/reorder", requestBody)

	router := mux.NewRouter()
	router.HandleFunc("/playlists/{uuid}/reorder", wrapWithAuth(t, handler.ReorderPlaylistTracks, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "invalid")
}
