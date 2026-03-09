//go:build integration

package builders

import (
	sqlhandler "backend/sql/sqlc"
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

// MusicBuilder provides a fluent interface for creating test music
type MusicBuilder struct {
	fromArtist        pgtype.UUID
	uploadedBy        pgtype.UUID
	inAlbum           pgtype.UUID
	songName          string
	pathInFileStorage string
	durationSeconds   int32
	imagePath         string
}

// NewMusicBuilder creates a new MusicBuilder with random defaults
func NewMusicBuilder(artistUUID, uploadedByUserUUID pgtype.UUID) *MusicBuilder {
	id := uuid.New().String()[:8]
	return &MusicBuilder{
		fromArtist:        artistUUID,
		uploadedBy:        uploadedByUserUUID,
		inAlbum:           pgtype.UUID{Valid: false}, // Nullable
		songName:          fmt.Sprintf("Song_%s", id),
		pathInFileStorage: fmt.Sprintf("audio/%s.mp3", uuid.New().String()),
		durationSeconds:   180, // 3 minutes default
		imagePath:         "",
	}
}

// WithSongName sets the song name
func (b *MusicBuilder) WithSongName(songName string) *MusicBuilder {
	b.songName = songName
	return b
}

// WithTitle is an alias for WithSongName for backward compatibility
func (b *MusicBuilder) WithTitle(title string) *MusicBuilder {
	return b.WithSongName(title)
}

// WithPathInFileStorage sets the file storage path
func (b *MusicBuilder) WithPathInFileStorage(path string) *MusicBuilder {
	b.pathInFileStorage = path
	return b
}

// WithDurationSeconds sets the duration
func (b *MusicBuilder) WithDurationSeconds(seconds int32) *MusicBuilder {
	b.durationSeconds = seconds
	return b
}

// WithImagePath sets the image path
func (b *MusicBuilder) WithImagePath(imagePath string) *MusicBuilder {
	b.imagePath = imagePath
	return b
}

// WithAlbum sets the album UUID
func (b *MusicBuilder) WithAlbum(albumUUID pgtype.UUID) *MusicBuilder {
	b.inAlbum = albumUUID
	return b
}

// Build creates the music in the database and returns the music UUID
func (b *MusicBuilder) Build(t *testing.T, ctx context.Context, db interface {
	CreateMusic(ctx context.Context, arg sqlhandler.CreateMusicParams) error
	GetMusicForArtist(ctx context.Context, arg sqlhandler.GetMusicForArtistParams) ([]sqlhandler.Music, error)
}) pgtype.UUID {
	t.Helper()

	err := db.CreateMusic(ctx, sqlhandler.CreateMusicParams{
		FromArtist:        b.fromArtist,
		UploadedBy:        b.uploadedBy,
		InAlbum:           b.inAlbum,
		SongName:          b.songName,
		PathInFileStorage: b.pathInFileStorage,
		DurationSeconds:   b.durationSeconds,
		ImagePath:         NewPgtypeText(b.imagePath),
	})

	require.NoError(t, err, "failed to create test music")

	// Query to get the music UUID that was just created
	music, err := db.GetMusicForArtist(ctx, sqlhandler.GetMusicForArtistParams{
		FromArtist: b.fromArtist,
		Limit:      100, // Get all music for artist
	})
	require.NoError(t, err, "failed to get music for artist")
	require.NotEmpty(t, music, "no music found for artist")

	// Return the UUID of the most recently created music (first one, query orders by created_at DESC)
	return music[0].Uuid
}

// BuildWithData returns the builder data without creating in DB
func (b *MusicBuilder) BuildWithData() sqlhandler.CreateMusicParams {
	return sqlhandler.CreateMusicParams{
		FromArtist:        b.fromArtist,
		UploadedBy:        b.uploadedBy,
		InAlbum:           b.inAlbum,
		SongName:          b.songName,
		PathInFileStorage: b.pathInFileStorage,
		DurationSeconds:   b.durationSeconds,
		ImagePath:         NewPgtypeText(b.imagePath),
	}
}
