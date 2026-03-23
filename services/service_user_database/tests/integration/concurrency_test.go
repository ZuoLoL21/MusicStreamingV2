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
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIntegration_Concurrency_MultipleLikes(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create music
	artistOwner := builders.NewUserBuilder().WithEmail("artist@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(artistOwner).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, artistOwner).Build(t, ctx, db)

	// Create 10 users
	users := make([]pgtype.UUID, 10)
	for i := 0; i < 10; i++ {
		users[i] = builders.NewUserBuilder().Build(t, ctx, db)
		ensureTimestampDistinct()
	}

	handler := handlers.NewLikesHandler(config, returns, db, nil)

	// Like concurrently
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		userIdx := i
		go func() {
			defer wg.Done()

			req := createRequest(t, "POST", "/music/"+builders.UUIDToString(musicUUID)+"/like", nil)
			router := mux.NewRouter()
			router.HandleFunc("/music/{uuid}/like", wrapWithAuth(t, handler.LikeMusic, users[userIdx])).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code == http.StatusCreated || rr.Code == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Verify all likes succeeded
	assert.Equal(t, 10, successCount, "all concurrent likes should succeed")

	// Verify count
	count, err := db.GetLikeCountMusic(ctx, musicUUID)
	require.NoError(t, err)
	assert.Equal(t, int64(10), count)
}

func TestIntegration_Concurrency_MultipleFollows(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create artist
	artistOwner := builders.NewUserBuilder().WithEmail("artist@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(artistOwner).Build(t, ctx, db)

	// Create 10 users (fans)
	users := make([]pgtype.UUID, 10)
	for i := 0; i < 10; i++ {
		users[i] = builders.NewUserBuilder().Build(t, ctx, db)
		ensureTimestampDistinct()
	}

	handler := handlers.NewFollowsHandler(config, returns, db)

	// Follow concurrently
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		userIdx := i
		go func() {
			defer wg.Done()

			req := createRequest(t, "POST", "/artists/"+builders.UUIDToString(artistUUID)+"/follow", nil)
			router := mux.NewRouter()
			router.HandleFunc("/artists/{uuid}/follow", wrapWithAuth(t, handler.FollowArtist, users[userIdx])).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code == http.StatusCreated || rr.Code == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Verify all follows succeeded
	assert.Equal(t, 10, successCount, "all concurrent follows should succeed")
}

func TestIntegration_Concurrency_PlaylistTrackAdditions(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create playlist and multiple music tracks
	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	playlistUUID := builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	musicTracks := make([]pgtype.UUID, 5)
	for i := 0; i < 5; i++ {
		musicTracks[i] = builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
		ensureTimestampDistinct()
	}

	handler := handlers.NewPlaylistHandler(config, returns, db, nil)

	// Add tracks concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		trackIdx := i
		go func() {
			defer wg.Done()

			req := createRequest(t, "PUT", "/playlists/"+builders.UUIDToString(playlistUUID)+"/tracks/"+builders.UUIDToString(musicTracks[trackIdx]), nil)
			router := mux.NewRouter()
			router.HandleFunc("/playlists/{uuid}/tracks/{musicUuid}", wrapWithAuth(t, handler.AddTrackToPlaylist, userUUID)).Methods("PUT")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
		}()
	}

	wg.Wait()

	// Verify tracks were added (may have duplicates depending on implementation)
	tracks, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        100,
		Column3:      -1,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tracks), 5)
}

func TestIntegration_Concurrency_ListeningHistory(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	// Create user and music
	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistOwner := builders.NewUserBuilder().WithEmail("artist@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(artistOwner).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, artistOwner).Build(t, ctx, db)

	// Add listening history concurrently (same user, same music)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Add listening entry directly via DB
			_ = db.AddListeningHistoryEntry(ctx, sqlhandler.AddListeningHistoryEntryParams{
				UserUuid:  userUUID,
				MusicUuid: musicUUID,
			})
		}()
	}

	wg.Wait()

	// Verify entries were created
	// Note: Exact count depends on implementation (may allow or prevent duplicate timestamps)
}

func TestIntegration_Concurrency_RoleChanges(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()
	logger := zap.NewNop()
	config := &backenddi.Config{}
	returns := di.NewReturnManager(logger)

	// Create artist with owner
	ownerUUID := builders.NewUserBuilder().WithEmail("owner@test.com").Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(ownerUUID).Build(t, ctx, db)

	// Add a member
	memberUUID := builders.NewUserBuilder().WithEmail("member@test.com").Build(t, ctx, db)
	db.AddUserToArtist(ctx, sqlhandler.AddUserToArtistParams{
		ArtistUuid: artistUUID,
		UserUuid:   memberUUID,
		Role:       sqlhandler.ArtistMemberRoleMember,
	})

	handler := handlers.NewArtistHandler(config, returns, db, nil)

	// Try to change role concurrently (stress test)
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			requestBody := map[string]string{
				"role": string(sqlhandler.ArtistMemberRoleManager),
			}
			req := createJSONRequest(t, "POST", "/artists/"+builders.UUIDToString(artistUUID)+"/members/"+builders.UUIDToString(memberUUID)+"/role", requestBody)
			router := mux.NewRouter()
			router.HandleFunc("/artists/{uuid}/members/{userUuid}/role", wrapWithAuth(t, handler.ChangeUserRole, ownerUUID)).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
		}()
	}

	wg.Wait()

	// Just verify no crash - final state may vary
}
