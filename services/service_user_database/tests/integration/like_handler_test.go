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

func TestIntegration_LikesHandler_LikeMusic_Success(t *testing.T) {
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

	handler := handlers.NewLikesHandler(logger, config, returns, db, fileStorage)

	req := createRequest(t, "POST", "/music/"+builders.UUIDToString(musicUUID)+"/like", nil)

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}/like", wrapWithAuth(t, handler.LikeMusic, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify like was created
	liked, err := db.IsLiked(ctx, sqlhandler.IsLikedParams{
		FromUser: userUUID,
		ToMusic:  musicUUID,
	})
	require.NoError(t, err)
	assert.True(t, liked)
}

func TestIntegration_LikesHandler_UnlikeMusic_Success(t *testing.T) {
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

	// First like the music
	err := db.LikeMusic(ctx, sqlhandler.LikeMusicParams{
		FromUser: userUUID,
		ToMusic:  musicUUID,
	})
	require.NoError(t, err)

	handler := handlers.NewLikesHandler(logger, config, returns, db, fileStorage)

	req := createRequest(t, "DELETE", "/music/"+builders.UUIDToString(musicUUID)+"/like", nil)

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}/like", wrapWithAuth(t, handler.UnlikeMusic, userUUID)).Methods("DELETE")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify like was removed
	liked, err := db.IsLiked(ctx, sqlhandler.IsLikedParams{
		FromUser: userUUID,
		ToMusic:  musicUUID,
	})
	require.NoError(t, err)
	assert.False(t, liked)
}

func TestIntegration_LikesHandler_IsLiked_True(t *testing.T) {
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

	// Like the music
	err := db.LikeMusic(ctx, sqlhandler.LikeMusicParams{
		FromUser: userUUID,
		ToMusic:  musicUUID,
	})
	require.NoError(t, err)

	handler := handlers.NewLikesHandler(logger, config, returns, db, fileStorage)

	req := createRequest(t, "GET", "/music/"+builders.UUIDToString(musicUUID)+"/liked", nil)

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}/liked", wrapWithAuth(t, handler.IsLiked, userUUID)).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]bool
	assertJSONResponse(t, rr, http.StatusOK, &response)
	assert.True(t, response["liked"])
}

func TestIntegration_LikesHandler_IsLiked_False(t *testing.T) {
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

	handler := handlers.NewLikesHandler(logger, config, returns, db, fileStorage)

	req := createRequest(t, "GET", "/music/"+builders.UUIDToString(musicUUID)+"/liked", nil)

	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}/liked", wrapWithAuth(t, handler.IsLiked, userUUID)).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]bool
	assertJSONResponse(t, rr, http.StatusOK, &response)
	assert.False(t, response["liked"])
}

func TestIntegration_LikesHandler_GetLikesForUser_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	// Create and like multiple tracks
	music1UUID := builders.NewMusicBuilder(artistUUID, userUUID).WithSongName("Liked Song 1").Build(t, ctx, db)
	ensureTimestampDistinct()
	music2UUID := builders.NewMusicBuilder(artistUUID, userUUID).WithSongName("Liked Song 2").Build(t, ctx, db)
	ensureTimestampDistinct()
	music3UUID := builders.NewMusicBuilder(artistUUID, userUUID).WithSongName("Liked Song 3").Build(t, ctx, db)

	err := db.LikeMusic(ctx, sqlhandler.LikeMusicParams{FromUser: userUUID, ToMusic: music1UUID})
	require.NoError(t, err)
	ensureTimestampDistinct()
	err = db.LikeMusic(ctx, sqlhandler.LikeMusicParams{FromUser: userUUID, ToMusic: music2UUID})
	require.NoError(t, err)
	ensureTimestampDistinct()
	err = db.LikeMusic(ctx, sqlhandler.LikeMusicParams{FromUser: userUUID, ToMusic: music3UUID})
	require.NoError(t, err)

	handler := handlers.NewLikesHandler(logger, config, returns, db, fileStorage)

	req := createRequest(t, "GET", "/users/"+builders.UUIDToString(userUUID)+"/likes", nil)

	router := mux.NewRouter()
	router.HandleFunc("/users/{uuid}/likes", handler.GetLikesForUser).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var likes []sqlhandler.Music
	assertJSONResponse(t, rr, http.StatusOK, &likes)
	assert.Len(t, likes, 3)
}

func TestIntegration_LikesHandler_GetLikeCountMusic_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	user1UUID := builders.NewUserBuilder().Build(t, ctx, db)
	user2UUID := builders.NewUserBuilder().WithEmail("user2@test.com").Build(t, ctx, db)
	user3UUID := builders.NewUserBuilder().WithEmail("user3@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(user1UUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, user1UUID).Build(t, ctx, db)

	// Multiple users like the same track
	err := db.LikeMusic(ctx, sqlhandler.LikeMusicParams{FromUser: user1UUID, ToMusic: musicUUID})
	require.NoError(t, err)
	err = db.LikeMusic(ctx, sqlhandler.LikeMusicParams{FromUser: user2UUID, ToMusic: musicUUID})
	require.NoError(t, err)
	err = db.LikeMusic(ctx, sqlhandler.LikeMusicParams{FromUser: user3UUID, ToMusic: musicUUID})
	require.NoError(t, err)

	// Verify count
	count, err := db.GetLikeCountMusic(ctx, musicUUID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestIntegration_LikesHandler_GetLikeCountUser_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	// User likes multiple tracks
	music1UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music2UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music3UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music4UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)

	err := db.LikeMusic(ctx, sqlhandler.LikeMusicParams{FromUser: userUUID, ToMusic: music1UUID})
	require.NoError(t, err)
	err = db.LikeMusic(ctx, sqlhandler.LikeMusicParams{FromUser: userUUID, ToMusic: music2UUID})
	require.NoError(t, err)
	err = db.LikeMusic(ctx, sqlhandler.LikeMusicParams{FromUser: userUUID, ToMusic: music3UUID})
	require.NoError(t, err)
	err = db.LikeMusic(ctx, sqlhandler.LikeMusicParams{FromUser: userUUID, ToMusic: music4UUID})
	require.NoError(t, err)

	// Verify count
	count, err := db.GetLikeCountUser(ctx, userUUID)
	require.NoError(t, err)
	assert.Equal(t, int64(4), count)
}

func TestIntegration_LikesHandler_LikeDuplicateIdempotent(t *testing.T) {
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

	handler := handlers.NewLikesHandler(logger, config, returns, db, fileStorage)

	req1 := createRequest(t, "POST", "/music/"+builders.UUIDToString(musicUUID)+"/like", nil)
	router := mux.NewRouter()
	router.HandleFunc("/music/{uuid}/like", wrapWithAuth(t, handler.LikeMusic, userUUID)).Methods("POST")

	// First like
	rr1 := httptest.NewRecorder()
	router.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Second like (duplicate) - should handle gracefully
	req2 := createRequest(t, "POST", "/music/"+builders.UUIDToString(musicUUID)+"/like", nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	// Depending on implementation, this might be 200 (idempotent) or 500/409
	// We verify count is still 1
	count, err := db.GetLikeCountMusic(ctx, musicUUID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
