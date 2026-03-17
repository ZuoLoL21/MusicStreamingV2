//go:build integration

package integration

import (
	sqlhandler "backend/sql/sqlc"
	"backend/tests/integration/builders"
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// ALBUM DATABASE TESTS
// ============================================================================

func TestIntegration_DB_Album_UpdateDetails(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	// Create artist and album
	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	builders.NewAlbumBuilder(artistUUID).
		WithName("Original Album").
		WithDescription("Original description").
		Build(t, ctx, db)

	// Get the created album
	albums, err := db.GetAlbumsForArtist(ctx, sqlhandler.GetAlbumsForArtistParams{
		FromArtist: artistUUID,
		Limit:      1,
	})
	require.NoError(t, err)
	require.Len(t, albums, 1)
	albumUUID := albums[0].Uuid

	// Update album
	err = db.UpdateAlbum(ctx, sqlhandler.UpdateAlbumParams{
		Uuid:         albumUUID,
		OriginalName: "Updated Album",
		Description:  pgtype.Text{String: "Updated description", Valid: true},
	})
	require.NoError(t, err)

	// Verify update
	updatedAlbums, err := db.GetAlbumsForArtist(ctx, sqlhandler.GetAlbumsForArtistParams{
		FromArtist: artistUUID,
		Limit:      1,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Album", updatedAlbums[0].OriginalName)
	assert.Equal(t, "Updated description", updatedAlbums[0].Description.String)
}

func TestIntegration_DB_Album_DeleteCascade(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	// Create artist and album
	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	builders.NewAlbumBuilder(artistUUID).Build(t, ctx, db)

	albums, err := db.GetAlbumsForArtist(ctx, sqlhandler.GetAlbumsForArtistParams{
		FromArtist: artistUUID,
		Limit:      1,
	})
	require.NoError(t, err)
	albumUUID := albums[0].Uuid

	// Create music in the album
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).
		WithAlbum(albumUUID).
		Build(t, ctx, db)

	// Verify music is in album
	music, err := db.GetMusicForAlbum(ctx, sqlhandler.GetMusicForAlbumParams{
		InAlbum: albumUUID,
		Limit:   10,
	})
	require.NoError(t, err)
	assert.Len(t, music, 1)
	assert.Equal(t, musicUUID, music[0].Uuid)

	// Delete album
	err = db.DeleteAlbum(ctx, albumUUID)
	require.NoError(t, err)

	// Verify music still exists but album_uuid is now NULL
	musicAfter, err := db.GetMusicForArtist(ctx, sqlhandler.GetMusicForArtistParams{
		FromArtist: artistUUID,
		Limit:      10,
	})
	require.NoError(t, err)
	assert.Len(t, musicAfter, 1)
	assert.False(t, musicAfter[0].InAlbum.Valid) // Album reference should be NULL
}

func TestIntegration_DB_Album_GetByArtistPagination(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	// Create multiple albums
	for i := 0; i < 5; i++ {
		builders.NewAlbumBuilder(artistUUID).Build(t, ctx, db)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get first page
	page1, err := db.GetAlbumsForArtist(ctx, sqlhandler.GetAlbumsForArtistParams{
		FromArtist: artistUUID,
		Limit:      2,
	})
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Get second page using cursor
	page2, err := db.GetAlbumsForArtist(ctx, sqlhandler.GetAlbumsForArtistParams{
		FromArtist: artistUUID,
		Limit:      2,
		Column3:    pgtype.Timestamptz{Time: page1[1].UpdatedAt.Time, Valid: page1[1].UpdatedAt.Valid},
		Uuid:       page1[1].Uuid,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(page2), 1)
	// Verify different albums returned
	assert.NotEqual(t, page1[0].Uuid, page2[0].Uuid)
}

func TestIntegration_DB_Album_MusicAssociation(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	builders.NewAlbumBuilder(artistUUID).WithName("Test Album").Build(t, ctx, db)

	albums, err := db.GetAlbumsForArtist(ctx, sqlhandler.GetAlbumsForArtistParams{
		FromArtist: artistUUID,
		Limit:      1,
	})
	require.NoError(t, err)
	albumUUID := albums[0].Uuid

	// Create music with album association
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).
		WithSongName("Album Track").
		WithAlbum(albumUUID).
		Build(t, ctx, db)

	// Verify music is in album
	albumMusic, err := db.GetMusicForAlbum(ctx, sqlhandler.GetMusicForAlbumParams{
		InAlbum: albumUUID,
		Limit:   10,
	})
	require.NoError(t, err)
	assert.Len(t, albumMusic, 1)
	assert.Equal(t, musicUUID, albumMusic[0].Uuid)
	assert.Equal(t, "Album Track", albumMusic[0].SongName)
}

func TestIntegration_DB_Album_MultiplePerArtist(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	// Create multiple albums for same artist
	albumNames := []string{"Album A", "Album B", "Album C"}
	for _, name := range albumNames {
		builders.NewAlbumBuilder(artistUUID).WithName(name).Build(t, ctx, db)
		time.Sleep(10 * time.Millisecond)
	}

	// Get all albums
	albums, err := db.GetAlbumsForArtist(ctx, sqlhandler.GetAlbumsForArtistParams{
		FromArtist: artistUUID,
		Limit:      10,
	})
	require.NoError(t, err)
	assert.Len(t, albums, 3)

	// Verify all names present
	foundNames := make(map[string]bool)
	for _, album := range albums {
		foundNames[album.OriginalName] = true
	}
	for _, name := range albumNames {
		assert.True(t, foundNames[name], "Album %s should exist", name)
	}
}

// ============================================================================
// PLAYLIST DATABASE TESTS
// ============================================================================

func TestIntegration_DB_Playlist_UpdateDetails(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).
		WithName("Original Playlist").
		WithPublic(true).
		Build(t, ctx, db)

	// Get created playlist
	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	// Update playlist (requires user UUID for authorization)
	err = db.UpdatePlaylist(ctx, sqlhandler.UpdatePlaylistParams{
		UserUuid:     userUUID,
		Uuid:         playlistUUID,
		OriginalName: "Updated Playlist",
		Description:  pgtype.Text{String: "New description", Valid: true},
		IsPublic:     pgtype.Bool{Bool: false, Valid: true},
	})
	require.NoError(t, err)

	// Verify update
	updated, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Playlist", updated[0].OriginalName)
	assert.Equal(t, "New description", updated[0].Description.String)
	assert.False(t, updated[0].IsPublic.Bool)
}

func TestIntegration_DB_Playlist_DeleteCascade(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	// Create user, artist, music, and playlist
	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	// Add track to playlist
	err = db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    musicUUID,
	})
	require.NoError(t, err)

	// Verify track exists
	tracks, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        10,
		Column3:      -1, // -1 means no filter (get all positions)
	})
	require.NoError(t, err)
	assert.Len(t, tracks, 1)

	// Delete playlist (requires user UUID for authorization)
	err = db.DeletePlaylist(ctx, sqlhandler.DeletePlaylistParams{
		UserUuid: userUUID,
		Uuid:     playlistUUID,
	})
	require.NoError(t, err)

	// Verify tracks are deleted (CASCADE)
	tracksAfter, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        10,
		Column3:      -1,
	})
	require.NoError(t, err)
	assert.Len(t, tracksAfter, 0)
}

func TestIntegration_DB_Playlist_AddRemoveTracks(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	music1UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music2UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	// Add first track
	err = db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music1UUID,
	})
	require.NoError(t, err)

	// Add second track
	err = db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music2UUID,
	})
	require.NoError(t, err)

	// Verify both tracks exist
	tracks, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        10,
		Column3:      -1,
	})
	require.NoError(t, err)
	require.Len(t, tracks, 2, "Should have 2 tracks after adding both")
	t.Logf("Track 1 UUID: %s, Track 2 UUID: %s", builders.UUIDToString(tracks[0].Uuid), builders.UUIDToString(tracks[1].Uuid))
	t.Logf("Music 1 UUID: %s, Music 2 UUID: %s", builders.UUIDToString(music1UUID), builders.UUIDToString(music2UUID))

	// Remove first track (requires user UUID for authorization)
	err = db.RemoveTrackFromPlaylist(ctx, sqlhandler.RemoveTrackFromPlaylistParams{
		UserUuid:     userUUID,
		PlaylistUuid: playlistUUID,
		MusicUuid:    music1UUID,
	})
	require.NoError(t, err)

	// Verify only one track remains
	tracksAfter, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        10,
		Column3:      -1,
	})
	require.NoError(t, err)
	if len(tracksAfter) > 0 {
		t.Logf("Remaining track UUID: %s", builders.UUIDToString(tracksAfter[0].Uuid))
	}
	require.Len(t, tracksAfter, 1, "Should have 1 track after removing one")
	assert.Equal(t, music2UUID, tracksAfter[0].Uuid)
}

func TestIntegration_DB_Playlist_ReorderTracks(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

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

	// Add tracks in order: music1, music2, music3 (positions 0, 1, 2)
	err = db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music1UUID,
	})
	require.NoError(t, err)
	err = db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music2UUID,
	})
	require.NoError(t, err)
	err = db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    music3UUID,
	})
	require.NoError(t, err)

	// Get initial order
	tracksBefore, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        10,
		Column3:      -1,
	})
	require.NoError(t, err)
	require.Len(t, tracksBefore, 3)
	assert.Equal(t, music1UUID, tracksBefore[0].Uuid) // position 0
	assert.Equal(t, music2UUID, tracksBefore[1].Uuid) // position 1
	assert.Equal(t, music3UUID, tracksBefore[2].Uuid) // position 2

	// Reorder tracks: music3, music1, music2 (reverse and swap)
	err = db.ReorderPlaylistTracks(ctx, sqlhandler.ReorderPlaylistTracksParams{
		UserUuid:     userUUID,
		PlaylistUuid: playlistUUID,
		Column3:      []pgtype.UUID{music3UUID, music1UUID, music2UUID},
	})
	require.NoError(t, err)

	// Get reordered tracks
	tracksAfter, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        10,
		Column3:      -1,
	})
	require.NoError(t, err)
	require.Len(t, tracksAfter, 3)
	assert.Equal(t, music3UUID, tracksAfter[0].Uuid) // now position 0
	assert.Equal(t, music1UUID, tracksAfter[1].Uuid) // now position 1
	assert.Equal(t, music2UUID, tracksAfter[2].Uuid) // now position 2
}

func TestIntegration_DB_Playlist_ReorderTracks_CountMismatch(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

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

	// Add 3 tracks
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

	// Try to reorder with only 2 tracks (count mismatch) - should fail
	err = db.ReorderPlaylistTracks(ctx, sqlhandler.ReorderPlaylistTracksParams{
		UserUuid:     userUUID,
		PlaylistUuid: playlistUUID,
		Column3:      []pgtype.UUID{music1UUID, music2UUID},
	})
	// Validation should prevent update - no error but no changes
	require.NoError(t, err)

	// Verify order unchanged
	tracks, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        10,
		Column3:      -1,
	})
	require.NoError(t, err)
	require.Len(t, tracks, 3)
	assert.Equal(t, music1UUID, tracks[0].Uuid)
	assert.Equal(t, music2UUID, tracks[1].Uuid)
	assert.Equal(t, music3UUID, tracks[2].Uuid)
}

func TestIntegration_DB_Playlist_ReorderTracks_InvalidTrack(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	music1UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music2UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	music3UUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	musicNotInPlaylistUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	// Add 3 tracks
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

	// Try to reorder with a track not in playlist - should fail
	err = db.ReorderPlaylistTracks(ctx, sqlhandler.ReorderPlaylistTracksParams{
		UserUuid:     userUUID,
		PlaylistUuid: playlistUUID,
		Column3:      []pgtype.UUID{music1UUID, music2UUID, musicNotInPlaylistUUID},
	})
	// Validation should prevent update
	require.NoError(t, err)

	// Verify order unchanged
	tracks, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        10,
		Column3:      -1,
	})
	require.NoError(t, err)
	require.Len(t, tracks, 3)
	assert.Equal(t, music1UUID, tracks[0].Uuid)
	assert.Equal(t, music2UUID, tracks[1].Uuid)
	assert.Equal(t, music3UUID, tracks[2].Uuid)
}

func TestIntegration_DB_Playlist_GetTracksWithPagination(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	// Add multiple tracks (positions auto-calculated)
	for i := 0; i < 5; i++ {
		musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
		db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
			PlaylistUuid: playlistUUID,
			MusicUuid:    musicUUID,
		})
	}

	// Get first page
	page1, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        2,
		Column3:      -1, // -1 means get all (first page)
	})
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Get second page using cursor (Column3 is position for this query)
	page2, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        2,
		Column3:      1, // Get positions > 1 (skip first 2 at positions 0 and 1)
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(page2), 1)
	assert.NotEqual(t, page1[0].Uuid, page2[0].Uuid)
}

func TestIntegration_DB_Playlist_DuplicateTracks(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	builders.NewPlaylistBuilder(userUUID).Build(t, ctx, db)

	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: userUUID,
		Limit:    1,
	})
	require.NoError(t, err)
	playlistUUID := playlists[0].Uuid

	// Add same music twice - should work, different positions
	err = db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    musicUUID,
	})
	require.NoError(t, err)

	err = db.AddTrackToPlaylist(ctx, sqlhandler.AddTrackToPlaylistParams{
		PlaylistUuid: playlistUUID,
		MusicUuid:    musicUUID,
	})
	require.NoError(t, err)

	// Verify both entries exist (same music at different positions)
	tracks, err := db.GetPlaylistTracks(ctx, sqlhandler.GetPlaylistTracksParams{
		PlaylistUuid: playlistUUID,
		Limit:        10,
		Column3:      -1,
	})
	require.NoError(t, err)
	assert.Len(t, tracks, 2)
	// Both should be the same music
	assert.Equal(t, musicUUID, tracks[0].Uuid)
	assert.Equal(t, musicUUID, tracks[1].Uuid)
}

// ============================================================================
// FOLLOW DATABASE TESTS
// ============================================================================

func TestIntegration_DB_Follow_ArtistOperations(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	user1UUID := builders.NewUserBuilder().Build(t, ctx, db)
	user2UUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(user2UUID).Build(t, ctx, db)

	// User1 follows artist
	err := db.FollowArtist(ctx, sqlhandler.FollowArtistParams{
		FromUser: user1UUID,
		ToArtist: artistUUID,
	})
	require.NoError(t, err)

	// Verify following
	isFollowing, err := db.IsFollowingArtist(ctx, sqlhandler.IsFollowingArtistParams{
		FromUser: user1UUID,
		ToArtist: artistUUID,
	})
	require.NoError(t, err)
	assert.True(t, isFollowing)

	// Unfollow artist
	err = db.UnfollowArtist(ctx, sqlhandler.UnfollowArtistParams{
		FromUser: user1UUID,
		ToArtist: artistUUID,
	})
	require.NoError(t, err)

	// Verify not following
	isFollowing, err = db.IsFollowingArtist(ctx, sqlhandler.IsFollowingArtistParams{
		FromUser: user1UUID,
		ToArtist: artistUUID,
	})
	require.NoError(t, err)
	assert.False(t, isFollowing)
}

func TestIntegration_DB_Follow_GetFollowersForArtist(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	ownerUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(ownerUUID).Build(t, ctx, db)

	// Create multiple followers
	followerCount := 5
	for i := 0; i < followerCount; i++ {
		userUUID := builders.NewUserBuilder().Build(t, ctx, db)
		err := db.FollowArtist(ctx, sqlhandler.FollowArtistParams{
			FromUser: userUUID,
			ToArtist: artistUUID,
		})
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Get first page
	page1, err := db.GetFollowersForArtist(ctx, sqlhandler.GetFollowersForArtistParams{
		ToArtist: artistUUID,
		Limit:    2,
	})
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Get second page
	page2, err := db.GetFollowersForArtist(ctx, sqlhandler.GetFollowersForArtistParams{
		ToArtist: artistUUID,
		Limit:    2,
		Column3:  pgtype.Timestamptz{Time: page1[1].CreatedAt.Time, Valid: page1[1].CreatedAt.Valid},
		Uuid:     page1[1].Uuid,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(page2), 1)
}

func TestIntegration_DB_Follow_GetFollowedArtistsPagination(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)

	// User follows multiple artists
	artistUUIDs := make([]pgtype.UUID, 5)
	for i := 0; i < 5; i++ {
		ownerUUID := builders.NewUserBuilder().Build(t, ctx, db)
		artistUUID := builders.NewArtistBuilder(ownerUUID).Build(t, ctx, db)
		artistUUIDs[i] = artistUUID
		db.FollowArtist(ctx, sqlhandler.FollowArtistParams{
			FromUser: userUUID,
			ToArtist: artistUUID,
		})
		time.Sleep(10 * time.Millisecond)
	}
	t.Logf("Created %d artists with UUIDs: %v", len(artistUUIDs), artistUUIDs)

	// Get first page
	page1, err := db.GetFollowedArtistsForUser(ctx, sqlhandler.GetFollowedArtistsForUserParams{
		FromUser: userUUID,
		Limit:    2,
	})
	require.NoError(t, err)
	t.Logf("Got %d artists in page1", len(page1))
	require.Len(t, page1, 2, "Should get 2 artists in first page")

	// Get second page
	page2, err := db.GetFollowedArtistsForUser(ctx, sqlhandler.GetFollowedArtistsForUserParams{
		FromUser: userUUID,
		Limit:    2,
		Column3:  pgtype.Timestamptz{Time: page1[1].CreatedAt.Time, Valid: page1[1].CreatedAt.Valid},
		Uuid:     page1[1].Uuid,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(page2), 1)
	assert.NotEqual(t, page1[0].Uuid, page2[0].Uuid)
}

func TestIntegration_DB_Follow_CountFollowers(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	ownerUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(ownerUUID).Build(t, ctx, db)

	// Initially no followers
	count, err := db.GetFollowerCountForUser(ctx, artistUUID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Add followers
	followerCount := 3
	for i := 0; i < followerCount; i++ {
		userUUID := builders.NewUserBuilder().Build(t, ctx, db)
		db.FollowArtist(ctx, sqlhandler.FollowArtistParams{
			FromUser: userUUID,
			ToArtist: artistUUID,
		})
	}

	// Verify count (note: GetFollowerCountForUser might not exist, using GetFollowingCountForArtist)
	count, err = db.GetFollowingCountForArtist(ctx, artistUUID)
	require.NoError(t, err)
	assert.Equal(t, int64(followerCount), count)
}

// ============================================================================
// MUSIC DATABASE TESTS
// ============================================================================

func TestIntegration_DB_Music_UpdateDetails(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).
		WithSongName("Original Song").
		WithDurationSeconds(180).
		Build(t, ctx, db)

	// Update music details (note: UpdateMusicDetails only updates song_name and in_album)
	err := db.UpdateMusicDetails(ctx, sqlhandler.UpdateMusicDetailsParams{
		Uuid:     musicUUID,
		SongName: "Updated Song",
		InAlbum:  pgtype.UUID{Valid: false}, // Keep album NULL
	})
	require.NoError(t, err)

	// Verify update
	music, err := db.GetMusicForArtist(ctx, sqlhandler.GetMusicForArtistParams{
		FromArtist: artistUUID,
		Limit:      1,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Song", music[0].SongName)
}

func TestIntegration_DB_Music_GetByAlbum(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	builders.NewAlbumBuilder(artistUUID).Build(t, ctx, db)

	albums, err := db.GetAlbumsForArtist(ctx, sqlhandler.GetAlbumsForArtistParams{
		FromArtist: artistUUID,
		Limit:      1,
	})
	require.NoError(t, err)
	albumUUID := albums[0].Uuid

	// Create multiple music tracks in album
	for i := 0; i < 3; i++ {
		builders.NewMusicBuilder(artistUUID, userUUID).
			WithAlbum(albumUUID).
			Build(t, ctx, db)
		time.Sleep(10 * time.Millisecond)
	}

	// Get first page
	page1, err := db.GetMusicForAlbum(ctx, sqlhandler.GetMusicForAlbumParams{
		InAlbum: albumUUID,
		Limit:   2,
	})
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Get second page
	page2, err := db.GetMusicForAlbum(ctx, sqlhandler.GetMusicForAlbumParams{
		InAlbum: albumUUID,
		Limit:   2,
		Column3: pgtype.Timestamptz{Time: page1[1].UpdatedAt.Time, Valid: page1[1].UpdatedAt.Valid},
		Uuid:    page1[1].Uuid,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(page2), 1)
}

func TestIntegration_DB_Music_GetByArtist(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	// Create music for artist
	musicCount := 3
	for i := 0; i < musicCount; i++ {
		builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
	}

	// Get all music for artist
	music, err := db.GetMusicForArtist(ctx, sqlhandler.GetMusicForArtistParams{
		FromArtist: artistUUID,
		Limit:      10,
	})
	require.NoError(t, err)
	assert.Len(t, music, musicCount)

	// Verify all belong to same artist
	for _, m := range music {
		assert.Equal(t, artistUUID, m.FromArtist)
	}
}

func TestIntegration_DB_Music_TagAssociations(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	// Create tags
	builders.NewTagBuilder().WithName("rock").Build(t, ctx, db)
	builders.NewTagBuilder().WithName("pop").Build(t, ctx, db)

	// Create music
	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)

	// Assign tags
	err := db.AssignTagToMusic(ctx, sqlhandler.AssignTagToMusicParams{
		MusicUuid: musicUUID,
		TagName:   "rock",
	})
	require.NoError(t, err)
	err = db.AssignTagToMusic(ctx, sqlhandler.AssignTagToMusicParams{
		MusicUuid: musicUUID,
		TagName:   "pop",
	})
	require.NoError(t, err)

	// Verify tags assigned
	tags, err := db.GetTagsForMusic(ctx, sqlhandler.GetTagsForMusicParams{
		MusicUuid: musicUUID,
		Limit:     10,
	})
	require.NoError(t, err)
	assert.Len(t, tags, 2)

	tagNames := make(map[string]bool)
	for _, tag := range tags {
		tagNames[tag.TagName] = true
	}
	assert.True(t, tagNames["rock"])
	assert.True(t, tagNames["pop"])
}

// ============================================================================
// TAG DATABASE TESTS
// ============================================================================

func TestIntegration_DB_Tag_GetMusicByTag(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	// Create tag
	builders.NewTagBuilder().WithName("electronic").Build(t, ctx, db)

	// Create multiple music tracks with tag
	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)

	for i := 0; i < 5; i++ {
		musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)
		db.AssignTagToMusic(ctx, sqlhandler.AssignTagToMusicParams{
			MusicUuid: musicUUID,
			TagName:   "electronic",
		})
		time.Sleep(10 * time.Millisecond)
	}

	// Get first page
	page1, err := db.GetMusicForTag(ctx, sqlhandler.GetMusicForTagParams{
		TagName: "electronic",
		Limit:   2,
	})
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Get second page
	page2, err := db.GetMusicForTag(ctx, sqlhandler.GetMusicForTagParams{
		TagName: "electronic",
		Limit:   2,
		Column3: pgtype.Timestamptz{Time: page1[1].UpdatedAt.Time, Valid: page1[1].UpdatedAt.Valid},
		Uuid:    page1[1].Uuid,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(page2), 1)
}

func TestIntegration_DB_Tag_GetAllTags(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	// Create multiple tags
	tagNames := []string{"blues", "country", "funk", "gospel", "hip-hop"}
	for _, name := range tagNames {
		builders.NewTagBuilder().WithName(name).Build(t, ctx, db)
	}

	// Get first page
	page1, err := db.GetAllTags(ctx, sqlhandler.GetAllTagsParams{
		Limit: 2,
	})
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Get second page using name cursor
	page2, err := db.GetAllTags(ctx, sqlhandler.GetAllTagsParams{
		Limit:   2,
		Column2: page1[1].TagName,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(page2), 1)
	assert.NotEqual(t, page1[0].TagName, page2[0].TagName)
}

func TestIntegration_DB_Tag_MusicWithMultipleTags(t *testing.T) {
	pool, db := SetupTestDB(t)
	defer CleanupTestData(t, pool)

	ctx := context.Background()

	// Create multiple tags
	tagNames := []string{"ambient", "chill", "downtempo", "lofi"}
	for _, name := range tagNames {
		builders.NewTagBuilder().WithName(name).Build(t, ctx, db)
	}

	// Create one music track
	userUUID := builders.NewUserBuilder().Build(t, ctx, db)
	artistUUID := builders.NewArtistBuilder(userUUID).Build(t, ctx, db)
	musicUUID := builders.NewMusicBuilder(artistUUID, userUUID).Build(t, ctx, db)

	// Assign all tags to the music
	for _, name := range tagNames {
		err := db.AssignTagToMusic(ctx, sqlhandler.AssignTagToMusicParams{
			MusicUuid: musicUUID,
			TagName:   name,
		})
		require.NoError(t, err)
	}

	// Verify all tags are assigned
	tags, err := db.GetTagsForMusic(ctx, sqlhandler.GetTagsForMusicParams{
		MusicUuid: musicUUID,
		Limit:     10,
	})
	require.NoError(t, err)
	assert.Len(t, tags, len(tagNames))

	foundTags := make(map[string]bool)
	for _, tag := range tags {
		foundTags[tag.TagName] = true
	}
	for _, name := range tagNames {
		assert.True(t, foundTags[name], "Tag %s should be assigned", name)
	}
}
