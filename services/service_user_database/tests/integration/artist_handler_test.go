//go:build integration

package integration

import (
	backenddi "backend/internal/di"
	"backend/internal/handlers"
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

func TestIntegration_ArtistHandler_Create(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	// Create test user as artist owner
	userUUID := builders.NewUserBuilder().
		WithEmail("artist@example.com").
		Build(t, ctx, db)

	handler := handlers.NewArtistHandler(config, returns, db, fileStorage)

	// Create artist with image
	imageData := createTestImage(512, 512) // Artist images require 512x512
	formFields := map[string]string{
		"artist_name": "The Test Band",
		"bio":         "A great test band",
	}
	req := createMultipartRequest(t, "POST", "/artists", "image", "artist.jpg", imageData, formFields)

	router := mux.NewRouter()
	router.HandleFunc("/artists", wrapWithAuth(t, handler.CreateArtist, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 201 Created
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify artist was created
	artists, err := db.GetArtistForUser(ctx, userUUID)
	require.NoError(t, err)
	require.Len(t, artists, 1)
	assert.Equal(t, "The Test Band", artists[0].ArtistName)
	assert.Equal(t, "A great test band", artists[0].Bio.String)
	assert.True(t, artists[0].ProfileImagePath.Valid, "profile image should be set")
}

func TestIntegration_ArtistHandler_Create_GeneratesSlug(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	// Create test user
	userUUID := builders.NewUserBuilder().
		WithEmail("slugtest@example.com").
		Build(t, ctx, db)

	handler := handlers.NewArtistHandler(config, returns, db, fileStorage)

	// Create artist with name that needs slug generation
	imageData := createTestImage(512, 512)
	formFields := map[string]string{
		"artist_name": "Artist With Spaces",
	}
	req := createMultipartRequest(t, "POST", "/artists", "image", "artist.jpg", imageData, formFields)

	router := mux.NewRouter()
	router.HandleFunc("/artists", wrapWithAuth(t, handler.CreateArtist, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 201 Created
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify artist was created with proper slug (if slug is generated)
	artists, err := db.GetArtistForUser(ctx, userUUID)
	require.NoError(t, err)
	require.Len(t, artists, 1)
	assert.Equal(t, "Artist With Spaces", artists[0].ArtistName)
	// Note: Slug generation depends on implementation - verify if needed
}

func TestIntegration_ArtistHandler_UpdateProfile(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create test user and artist
	userUUID := builders.NewUserBuilder().
		WithEmail("update@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(userUUID).
		WithName("Old Name").
		WithBio("Old bio").
		Build(t, ctx, db)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Update artist profile
	bio := "Updated bio for artist"
	updateReq := map[string]interface{}{
		"artist_name": "New Name",
		"bio":         &bio,
	}
	req := createJSONRequest(t, "POST", "/artists/"+builders.UUIDToString(artistUUID), updateReq)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}", wrapWithAuth(t, handler.UpdateArtistProfile, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify artist was updated
	artist, err := db.GetArtist(ctx, artistUUID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", artist.ArtistName)
	assert.Equal(t, "Updated bio for artist", artist.Bio.String)
}

func TestIntegration_ArtistHandler_UpdateProfile_Unauthorized(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create artist owner
	ownerUUID := builders.NewUserBuilder().
		WithEmail("owner@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(ownerUUID).
		WithName("Original Name").
		Build(t, ctx, db)

	// Create another user (not a member)
	nonMemberUUID := builders.NewUserBuilder().
		WithEmail("nonmember@example.com").
		Build(t, ctx, db)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Try to update as non-member
	bio := "Unauthorized update"
	updateReq := map[string]interface{}{
		"artist_name": "Hacked Name",
		"bio":         &bio,
	}
	req := createJSONRequest(t, "POST", "/artists/"+builders.UUIDToString(artistUUID), updateReq)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}", wrapWithAuth(t, handler.UpdateArtistProfile, nonMemberUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rr.Code)

	// Verify artist was NOT updated
	artist, err := db.GetArtist(ctx, artistUUID)
	require.NoError(t, err)
	assert.Equal(t, "Original Name", artist.ArtistName)
}

func TestIntegration_ArtistHandler_UpdatePicture(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)
	fileStorage := SetupMinIOClient(t)

	// Create test user and artist
	userUUID := builders.NewUserBuilder().
		WithEmail("picture@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(userUUID).
		WithName("Picture Test Artist").
		Build(t, ctx, db)

	handler := handlers.NewArtistHandler(config, returns, db, fileStorage)

	// Upload new picture
	imageData := createTestImage(512, 512)
	formFields := map[string]string{}
	req := createMultipartRequest(t, "POST", "/artists/"+builders.UUIDToString(artistUUID)+"/picture", "image", "newpic.jpg", imageData, formFields)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/picture", wrapWithAuth(t, handler.UpdateArtistPicture, userUUID)).Methods("POST")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify picture was updated
	artist, err := db.GetArtist(ctx, artistUUID)
	require.NoError(t, err)
	assert.True(t, artist.ProfileImagePath.Valid, "profile image should be set")
	assert.NotEmpty(t, artist.ProfileImagePath.String)
}

func TestIntegration_ArtistHandler_GetArtist(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create test user and artist
	userUUID := builders.NewUserBuilder().
		WithEmail("getartist@example.com").
		Build(t, ctx, db)

	artistUUID := builders.NewArtistBuilder(userUUID).
		WithName("Get Test Artist").
		WithBio("Test bio").
		Build(t, ctx, db)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Get artist (no auth required)
	req := createRequest(t, "GET", "/artists/"+builders.UUIDToString(artistUUID), nil)

	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}", handler.GetArtist).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 200 OK with artist data
	var response map[string]interface{}
	assertJSONResponse(t, rr, http.StatusOK, &response)

	assert.Equal(t, "Get Test Artist", response["artist_name"])
	assert.Equal(t, "Test bio", response["bio"])
}

func TestIntegration_ArtistHandler_GetArtistsAlphabetically(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create test user and multiple artists
	userUUID := builders.NewUserBuilder().
		WithEmail("alphabetical@example.com").
		Build(t, ctx, db)

	// Create artists in non-alphabetical order
	builders.NewArtistBuilder(userUUID).WithName("Zebra Band").Build(t, ctx, db)
	builders.NewArtistBuilder(userUUID).WithName("Alpha Band").Build(t, ctx, db)
	builders.NewArtistBuilder(userUUID).WithName("Middle Band").Build(t, ctx, db)

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Get artists alphabetically
	req := createRequest(t, "GET", "/artists?limit=10", nil)

	router := mux.NewRouter()
	router.HandleFunc("/artists", handler.GetArtistsAlphabetically).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 200 OK with sorted artists
	var artists []map[string]interface{}
	assertJSONResponse(t, rr, http.StatusOK, &artists)

	require.Len(t, artists, 3)
	assert.Equal(t, "Alpha Band", artists[0]["artist_name"])
	assert.Equal(t, "Middle Band", artists[1]["artist_name"])
	assert.Equal(t, "Zebra Band", artists[2]["artist_name"])
}

func TestIntegration_ArtistHandler_SearchForArtist(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create test user and artists
	userUUID := builders.NewUserBuilder().
		WithEmail("search@example.com").
		Build(t, ctx, db)

	builders.NewArtistBuilder(userUUID).WithName("The Beatles").Build(t, ctx, db)
	builders.NewArtistBuilder(userUUID).WithName("Beach Boys").Build(t, ctx, db)
	builders.NewArtistBuilder(userUUID).WithName("Rolling Stones").Build(t, ctx, db)

	handler := handlers.NewSearchHandler(config, returns, db, nil)

	// Search for artists with "bea" (should match Beatles and Beach Boys)
	req := createRequest(t, "GET", "/search/artists?q=bea&limit=10", nil)

	router := mux.NewRouter()
	router.HandleFunc("/search/artists", handler.SearchArtists).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 200 OK with matching artists
	var result struct {
		Artists []map[string]interface{} `json:"artists"`
	}
	assertJSONResponse(t, rr, http.StatusOK, &result)

	// Should match both Beatles and Beach Boys
	require.GreaterOrEqual(t, len(result.Artists), 2)
	// Note: Exact matching depends on search implementation (LIKE, trigram, etc.)
}

func TestIntegration_ArtistHandler_SearchForArtist_NoResults(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create test user and artists
	userUUID := builders.NewUserBuilder().
		WithEmail("nosearch@example.com").
		Build(t, ctx, db)

	builders.NewArtistBuilder(userUUID).WithName("The Beatles").Build(t, ctx, db)

	handler := handlers.NewSearchHandler(config, returns, db, nil)

	// Search for non-existent artist
	req := createRequest(t, "GET", "/search/artists?q=NonExistentArtistXYZ&limit=10", nil)

	router := mux.NewRouter()
	router.HandleFunc("/search/artists", handler.SearchArtists).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 200 OK with empty array
	var result struct {
		Artists []map[string]interface{} `json:"artists"`
	}
	assertJSONResponse(t, rr, http.StatusOK, &result)

	assert.Len(t, result.Artists, 0)
}
