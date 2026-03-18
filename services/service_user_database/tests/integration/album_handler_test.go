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

func TestIntegration_AlbumHandler_CreateAlbum_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	// Setup: Create user and artist
	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	handler := handlers.NewAlbumHandler(config, returns, db, fileStorage)

	// Create album with image
	imageData := createTestImage(1024, 1024) // Album images require 1024x1024
	formFields := map[string]string{
		"artist_uuid":   builders.UUIDToString(artistUUID),
		"original_name": "Test Album",
		"description":   "A great test album",
	}
	req := createMultipartRequest(t, "POST", "/albums", "image", "album.jpg", imageData, formFields)

	router := mux.NewRouter()
	router.HandleFunc("/albums", wrapWithAuth(t, handler.CreateAlbum, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify album was created
	albums, err := db.GetAlbumsForArtist(ctx, sqlhandler.GetAlbumsForArtistParams{
		FromArtist: artistUUID,
		Limit:      100,
	})
	require.NoError(t, err)
	require.Len(t, albums, 1)
	assert.Equal(t, "Test Album", albums[0].OriginalName)
	assert.Equal(t, "A great test album", albums[0].Description.String)
	assert.True(t, albums[0].ImagePath.Valid)
}

func TestIntegration_AlbumHandler_CreateAlbum_WithoutImage(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	handler := handlers.NewAlbumHandler(config, returns, db, fileStorage)

	// Create album without image
	formFields := map[string]string{
		"artist_uuid":   builders.UUIDToString(artistUUID),
		"original_name": "Album No Image",
		"description":   "Album without cover",
	}
	req := createMultipartRequest(t, "POST", "/albums", "", "", nil, formFields)

	router := mux.NewRouter()
	router.HandleFunc("/albums", wrapWithAuth(t, handler.CreateAlbum, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestIntegration_AlbumHandler_CreateAlbum_Unauthorized(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	// Create two users - one owns artist, the other tries to create album
	ownerUUID := builders.NewUserBuilder().Build(t, ctx, db)
	nonMemberUUID := builders.NewUserBuilder().WithEmail("nonmember@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(ownerUUID).Build(t, ctx, db)

	handler := handlers.NewAlbumHandler(config, returns, db, fileStorage)

	formFields := map[string]string{
		"artist_uuid":   builders.UUIDToString(artistUUID),
		"original_name": "Unauthorized Album",
	}
	req := createMultipartRequest(t, "POST", "/albums", "", "", nil, formFields)

	router := mux.NewRouter()
	router.HandleFunc("/albums", wrapWithAuth(t, handler.CreateAlbum, nonMemberUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestIntegration_AlbumHandler_UpdateAlbum_Success(t *testing.T) {
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

	handler := handlers.NewAlbumHandler(config, returns, db, fileStorage)

	updateBody := map[string]interface{}{
		"original_name": "Updated Album Name",
		"description":   "Updated description",
	}
	req := createJSONRequest(t, "PUT", "/albums/"+builders.UUIDToString(albumUUID), updateBody)

	router := mux.NewRouter()
	router.HandleFunc("/albums/{uuid}", wrapWithAuth(t, handler.UpdateAlbum, userUUID)).Methods("PUT")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify update
	album, err := db.GetAlbum(ctx, albumUUID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Album Name", album.OriginalName)
	assert.Equal(t, "Updated description", album.Description.String)
}

func TestIntegration_AlbumHandler_UpdateAlbum_RequiresManagerRole(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	ownerUUID := builders.NewUserBuilder().Build(t, ctx, db)
	memberUUID := builders.NewUserBuilder().WithEmail("member@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(ownerUUID).Build(t, ctx, db)
	albumUUID := builders.NewAlbumBuilder(artistUUID).Build(t, ctx, db)

	// Add member with member role (not manager)
	err := db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		UserUuid:   memberUUID,
		ArtistUuid: artistUUID,
		Role:       sqlhandler.ArtistMemberRoleMember,
	})
	require.NoError(t, err)

	handler := handlers.NewAlbumHandler(config, returns, db, fileStorage)

	updateBody := map[string]interface{}{
		"original_name": "Trying to Update",
	}
	req := createJSONRequest(t, "PUT", "/albums/"+builders.UUIDToString(albumUUID), updateBody)

	router := mux.NewRouter()
	router.HandleFunc("/albums/{uuid}", wrapWithAuth(t, handler.UpdateAlbum, memberUUID)).Methods("PUT")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestIntegration_AlbumHandler_UpdateAlbumImage_Success(t *testing.T) {
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

	handler := handlers.NewAlbumHandler(config, returns, db, fileStorage)

	imageData := createTestImage(1024, 1024)
	req := createMultipartRequest(t, "PUT", "/albums/"+builders.UUIDToString(albumUUID)+"/image",
		"image", "newalbum.jpg", imageData, map[string]string{})

	router := mux.NewRouter()
	router.HandleFunc("/albums/{uuid}/image", wrapWithAuth(t, handler.UpdateAlbumImage, userUUID)).Methods("PUT")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify image path was updated
	album, err := db.GetAlbum(ctx, albumUUID)
	require.NoError(t, err)
	assert.True(t, album.ImagePath.Valid)
}

func TestIntegration_AlbumHandler_DeleteAlbum_Success(t *testing.T) {
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

	handler := handlers.NewAlbumHandler(config, returns, db, fileStorage)

	req := createRequest(t, "DELETE", "/albums/"+builders.UUIDToString(albumUUID), nil)

	router := mux.NewRouter()
	router.HandleFunc("/albums/{uuid}", wrapWithAuth(t, handler.DeleteAlbum, userUUID)).Methods("DELETE")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify album was deleted
	_, err := db.GetAlbum(ctx, albumUUID)
	assert.Error(t, err)
}

func TestIntegration_AlbumHandler_DeleteAlbum_RequiresOwnerRole(t *testing.T) {
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
	albumUUID := builders.NewAlbumBuilder(artistUUID).Build(t, ctx, db)

	// Add manager role (not owner)
	err := db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		UserUuid:   managerUUID,
		ArtistUuid: artistUUID,
		Role:       sqlhandler.ArtistMemberRoleManager,
	})
	require.NoError(t, err)

	handler := handlers.NewAlbumHandler(config, returns, db, fileStorage)

	req := createRequest(t, "DELETE", "/albums/"+builders.UUIDToString(albumUUID), nil)

	router := mux.NewRouter()
	router.HandleFunc("/albums/{uuid}", wrapWithAuth(t, handler.DeleteAlbum, managerUUID)).Methods("DELETE")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestIntegration_AlbumHandler_GetAlbum_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	albumUUID := builders.NewAlbumBuilder(artistUUID).WithName("Get Test Album").Build(t, ctx, db)

	handler := handlers.NewAlbumHandler(config, returns, db, fileStorage)

	req := createRequest(t, "GET", "/albums/"+builders.UUIDToString(albumUUID), nil)

	router := mux.NewRouter()
	router.HandleFunc("/albums/{uuid}", handler.GetAlbum).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var album sqlhandler.Album
	assertJSONResponse(t, rr, http.StatusOK, &album)
	assert.Equal(t, "Get Test Album", album.OriginalName)
}

func TestIntegration_AlbumHandler_GetAlbumsForArtist_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	// Create multiple albums
	builders.NewAlbumBuilder(artistUUID).WithName("Album 1").Build(t, ctx, db)
	ensureTimestampDistinct()
	builders.NewAlbumBuilder(artistUUID).WithName("Album 2").Build(t, ctx, db)
	ensureTimestampDistinct()
	builders.NewAlbumBuilder(artistUUID).WithName("Album 3").Build(t, ctx, db)

	handler := handlers.NewAlbumHandler(config, returns, db, fileStorage)

	req := createRequest(t, "GET", "/artists/"+builders.UUIDToString(artistUUID)+"/albums", nil)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/albums", handler.GetAlbumsForArtist).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var albums []sqlhandler.Album
	assertJSONResponse(t, rr, http.StatusOK, &albums)
	assert.Len(t, albums, 3)
}

func TestIntegration_AlbumHandler_SearchForAlbum_Success(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	// Create albums with searchable names
	builders.NewAlbumBuilder(artistUUID).WithName("Dark Side of the Moon").Build(t, ctx, db)
	builders.NewAlbumBuilder(artistUUID).WithName("The Wall").Build(t, ctx, db)
	builders.NewAlbumBuilder(artistUUID).WithName("Wish You Were Here").Build(t, ctx, db)

	handler := handlers.NewSearchHandler(config, returns, db, fileStorage)

	req := createRequest(t, "GET", "/search/albums?q=Dark%20Side", nil)

	router := mux.NewRouter()
	router.HandleFunc("/search/albums", handler.SearchAlbums).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// Search results should contain at least the album with "Moon" in the name
	assert.Contains(t, rr.Body.String(), "Dark Side of the Moon")
}
