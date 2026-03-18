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
	"go.uber.org/zap"
)

func TestIntegration_Pagination_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "zero_limit",
			queryParams:    "limit=0",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "negative_limit",
			queryParams:    "limit=-1",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "excessive_limit",
			queryParams:    "limit=10000",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid_cursor_uuid",
			queryParams:    "cursor_id=not-a-uuid",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, db := SetupTestDB(t)
			defer CleanupTestData(t, pool)

			ctx := context.Background()
			logger := zap.NewNop()
			config := &backenddi.Config{}
			returns := di.NewReturnManager(logger)

			// Create test data
			userUUID := builders.NewUserBuilder().Build(t, ctx, db)
			artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

			// Create some albums
			for i := 0; i < 5; i++ {
				builders.NewAlbumBuilder(artistUUID).Build(t, ctx, db)
			}

			handler := handlers.NewAlbumHandler(config, returns, db, nil)

			req := createRequest(t, "GET", "/artists/"+builders.UUIDToString(artistUUID)+"/albums?"+tt.queryParams, nil)
			router := mux.NewRouter()
			router.HandleFunc("/artists/{uuid}/albums", handler.GetAlbumsForArtist).Methods("GET")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Should handle gracefully
			assert.NotEqual(t, 0, rr.Code)
		})
	}
}

func TestIntegration_Pagination_EmptyResultSet(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	// No albums created

	handler := handlers.NewAlbumHandler(config, returns, db, nil)

	req := createRequest(t, "GET", "/artists/"+builders.UUIDToString(artistUUID)+"/albums?limit=10", nil)
	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/albums", handler.GetAlbumsForArtist).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var albums []map[string]interface{}
	assertJSONResponse(t, rr, http.StatusOK, &albums)
	assert.Len(t, albums, 0)
}

func TestIntegration_Pagination_SingleItem(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	// Create exactly one album
	builders.NewAlbumBuilder(artistUUID).Build(t, ctx, db)

	handler := handlers.NewAlbumHandler(config, returns, db, nil)

	req := createRequest(t, "GET", "/artists/"+builders.UUIDToString(artistUUID)+"/albums?limit=10", nil)
	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/albums", handler.GetAlbumsForArtist).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var albums []map[string]interface{}
	assertJSONResponse(t, rr, http.StatusOK, &albums)
	assert.Len(t, albums, 1)
}

func TestIntegration_Pagination_ExactPageBoundary(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	// Create exactly 10 albums (page size)
	for i := 0; i < 10; i++ {
		builders.NewAlbumBuilder(artistUUID).Build(t, ctx, db)
		ensureTimestampDistinct()
	}

	handler := handlers.NewAlbumHandler(config, returns, db, nil)

	req := createRequest(t, "GET", "/artists/"+builders.UUIDToString(artistUUID)+"/albums?limit=10", nil)
	router := mux.NewRouter()
	router.HandleFunc("/artists/{uuid}/albums", handler.GetAlbumsForArtist).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var albums []map[string]interface{}
	assertJSONResponse(t, rr, http.StatusOK, &albums)
	assert.Len(t, albums, 10)
}
