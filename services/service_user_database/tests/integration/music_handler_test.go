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

func TestIntegration_MusicHandler_CreateMusic_WithAudio(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	audioData := []byte("fake audio data")
	formFields := map[string]string{
		"artist_uuid":      builders.UUIDToString(artistUUID),
		"song_name":        "Test Song",
		"duration_seconds": "180",
	}

	// Create multipart request with both audio and image
	req := createMultipartRequest(t, "POST", "/music", "audio", "song.mp3", audioData, formFields)

	// Note: createMultipartRequest only supports one file, so we test with audio only
	// In production, both audio and image can be uploaded in one request

	router := mux.NewRouter()
	router.HandleFunc("/music", wrapWithAuth(t, handler.CreateMusic, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify music was created
	music, err := db.GetMusicForArtist(ctx, sqlhandler.GetMusicForArtistParams{
		FromArtist: artistUUID,
		Limit:      100,
	})
	require.NoError(t, err)
	require.Len(t, music, 1)
	assert.Equal(t, "Test Song", music[0].SongName)
	assert.Equal(t, int32(180), music[0].DurationSeconds)
	assert.NotEmpty(t, music[0].PathInFileStorage)
}

func TestIntegration_MusicHandler_CreateMusic_WithAlbum(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	albumUUID := builders.NewAlbumBuilder(artistUUID).Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	audioData := []byte("fake audio data")
	formFields := map[string]string{
		"artist_uuid":      builders.UUIDToString(artistUUID),
		"song_name":        "Album Song",
		"duration_seconds": "200",
		"in_album":         builders.UUIDToString(albumUUID),
	}
	req := createMultipartRequest(t, "POST", "/music", "audio", "song.mp3", audioData, formFields)

	router := mux.NewRouter()
	router.HandleFunc("/music", wrapWithAuth(t, handler.CreateMusic, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify music is linked to album
	music, err := db.GetMusicForAlbum(ctx, sqlhandler.GetMusicForAlbumParams{
		InAlbum: albumUUID,
		Limit:   100,
	})
	require.NoError(t, err)
	require.Len(t, music, 1)
	assert.Equal(t, "Album Song", music[0].SongName)
}

func TestIntegration_MusicHandler_CreateMusic_WithoutAlbum(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	audioData := []byte("fake audio data")
	formFields := map[string]string{
		"artist_uuid":      builders.UUIDToString(artistUUID),
		"song_name":        "Single Track",
		"duration_seconds": "240",
	}
	req := createMultipartRequest(t, "POST", "/music", "audio", "single.mp3", audioData, formFields)

	router := mux.NewRouter()
	router.HandleFunc("/music", wrapWithAuth(t, handler.CreateMusic, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestIntegration_MusicHandler_CreateMusic_Unauthorized(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	ownerUUID := builders.NewUserBuilder().Build(t, ctx, db)
	nonMemberUUID := builders.NewUserBuilder().WithEmail("nonmember@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(ownerUUID).Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	audioData := []byte("fake audio data")
	formFields := map[string]string{
		"artist_uuid":      builders.UUIDToString(artistUUID),
		"song_name":        "Unauthorized Song",
		"duration_seconds": "180",
	}
	req := createMultipartRequest(t, "POST", "/music", "audio", "song.mp3", audioData, formFields)

	router := mux.NewRouter()
	router.HandleFunc("/music", wrapWithAuth(t, handler.CreateMusic, nonMemberUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestIntegration_MusicHandler_UpdateMusicDetails_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	updateBody := map[string]interface{}{
		"song_name": "Updated Song Title",
	}
	req := createJSONRequest(t, "PUT", "/music/"+builders.UUIDToString(musicUUID), updateBody)

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}", wrapWithAuth(t, handler.UpdateMusicDetails, userUUID)).Methods("PUT")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify update
	music, err := db.GetMusic(ctx, musicUUID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Song Title", music.SongName)
}

func TestIntegration_MusicHandler_UpdateMusicDetails_WithAlbum(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	albumUUID := builders.NewAlbumBuilder(artistUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	albumStr := builders.UUIDToString(albumUUID)
	updateBody := map[string]interface{}{
		"song_name": "Song Now In Album",
		"in_album":  albumStr,
	}
	req := createJSONRequest(t, "PUT", "/music/"+builders.UUIDToString(musicUUID), updateBody)

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}", wrapWithAuth(t, handler.UpdateMusicDetails, userUUID)).Methods("PUT")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify music is now in album
	music, err := db.GetMusic(ctx, musicUUID)
	require.NoError(t, err)
	assert.True(t, music.InAlbum.Valid)
	assert.Equal(t, albumUUID.Bytes, music.InAlbum.Bytes)
}

func TestIntegration_MusicHandler_UpdateMusicImage_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	imageData := createTestImage(1024, 1024)
	req := createMultipartRequest(t, "PUT", "/music/"+builders.UUIDToString(musicUUID)+"/image",
		"image", "newcover.jpg", imageData, map[string]string{})

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}/image", wrapWithAuth(t, handler.UpdateMusicImage, userUUID)).Methods("PUT")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify image path was updated
	music, err := db.GetMusic(ctx, musicUUID)
	require.NoError(t, err)
	assert.True(t, music.ImagePath.Valid)
}

func TestIntegration_MusicHandler_UpdateMusicStorage_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	newAudioData := []byte("updated audio data")
	formFields := map[string]string{
		"duration_seconds": "250",
	}
	req := createMultipartRequest(t, "PUT", "/music/"+builders.UUIDToString(musicUUID)+"/storage",
		"audio", "newsong.mp3", newAudioData, formFields)

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}/storage", wrapWithAuth(t, handler.UpdateMusicStorage, userUUID)).Methods("PUT")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify audio and duration were updated
	music, err := db.GetMusic(ctx, musicUUID)
	require.NoError(t, err)
	assert.Equal(t, int32(250), music.DurationSeconds)
}

func TestIntegration_MusicHandler_DeleteMusic_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	req := createRequest(t, "DELETE", "/music/"+builders.UUIDToString(musicUUID), nil)

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}", wrapWithAuth(t, handler.DeleteMusic, userUUID)).Methods("DELETE")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify music was deleted
	_, err := db.GetMusic(ctx, musicUUID)
	assert.Error(t, err)
}

func TestIntegration_MusicHandler_DeleteMusic_RequiresOwnerRole(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	ownerUUID := builders.NewUserBuilder().Build(t, ctx, db)
	managerUUID := builders.NewUserBuilder().WithEmail("manager@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(ownerUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, ownerUUID).Build(t, ctx, db)

	// Add manager role (not owner)
	err := db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		UserUuid:   managerUUID,
		ArtistUuid: artistUUID,
		Role:       sqlhandler.ArtistMemberRoleManager,
	})
	require.NoError(t, err)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	req := createRequest(t, "DELETE", "/music/"+builders.UUIDToString(musicUUID), nil)

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}", wrapWithAuth(t, handler.DeleteMusic, managerUUID)).Methods("DELETE")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestIntegration_MusicHandler_IncrementPlayCount_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	req := createRequest(t, "POST", "/music/"+builders.UUIDToString(musicUUID)+"/play", nil)

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}/play", handler.IncrementPlayCount).Methods("POST")

	// Get initial play count
	initialMusic, err := db.GetMusic(ctx, musicUUID)
	require.NoError(t, err)
	initialCount := initialMusic.PlayCount.Int32

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify play count was incremented
	updatedMusic, err := db.GetMusic(ctx, musicUUID)
	require.NoError(t, err)
	assert.Equal(t, initialCount+1, updatedMusic.PlayCount.Int32)
}

func TestIntegration_MusicHandler_GetMusic_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).
		WithSongName("Get Test Song").
		Build(t, ctx, db)

	handler := handlers.NewMusicHandler(logger, config, returns, db, fileStorage)

	req := createRequest(t, "GET", "/music/"+builders.UUIDToString(musicUUID), nil)

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}", handler.GetMusic).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var music sqlhandler.Music
	assertJSONResponse(t, rr, http.StatusOK, &music)
	assert.Equal(t, "Get Test Song", music.SongName)
}

func TestIntegration_MusicHandler_SearchForMusic_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	// Create music with searchable names
	builders.NewMusicBuilder(artistUUID, userUUID).WithSongName("Bohemian Rhapsody").Build(t, ctx, db)
	builders.NewMusicBuilder(artistUUID, userUUID).WithSongName("Stairway to Heaven").Build(t, ctx, db)
	builders.NewMusicBuilder(artistUUID, userUUID).WithSongName("Hotel California").Build(t, ctx, db)

	handler := handlers.NewSearchHandler(logger, config, returns, db, fileStorage)

	req := createRequest(t, "GET", "/search/music?q=Bohemian", nil)

	router := mux.NewRouter()
	router.HandleFunc("/search/music", handler.SearchMusic).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// Search results should contain the song with "Bohemian" in the name
	assert.Contains(t, rr.Body.String(), "Bohemian Rhapsody")
}
