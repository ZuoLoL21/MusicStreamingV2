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

// AlbumBuilder provides a fluent interface for creating test albums
type AlbumBuilder struct {
	name        string
	description pgtype.Text
	imagePath   pgtype.Text
	artistUUID  pgtype.UUID
}

// NewAlbumBuilder creates a new AlbumBuilder with random defaults
func NewAlbumBuilder(artistUUID pgtype.UUID) *AlbumBuilder {
	id := uuid.New().String()[:8]
	return &AlbumBuilder{
		name:        fmt.Sprintf("Album_%s", id),
		description: NewPgtypeText("Test album description"),
		imagePath:   NewPgtypeText(""),
		artistUUID:  artistUUID,
	}
}

// WithName sets the album name
func (b *AlbumBuilder) WithName(name string) *AlbumBuilder {
	b.name = name
	return b
}

// WithDescription sets the description
func (b *AlbumBuilder) WithDescription(description string) *AlbumBuilder {
	b.description = NewPgtypeText(description)
	return b
}

// WithImagePath sets the image path
func (b *AlbumBuilder) WithImagePath(imagePath string) *AlbumBuilder {
	b.imagePath = NewPgtypeText(imagePath)
	return b
}

// Build creates the album in the database and returns the album UUID
func (b *AlbumBuilder) Build(t *testing.T, ctx context.Context, db interface {
	CreateAlbum(ctx context.Context, arg sqlhandler.CreateAlbumParams) error
	GetAlbumsForArtist(ctx context.Context, arg sqlhandler.GetAlbumsForArtistParams) ([]sqlhandler.Album, error)
}) pgtype.UUID {
	t.Helper()

	err := db.CreateAlbum(ctx, sqlhandler.CreateAlbumParams{
		FromArtist:   b.artistUUID,
		OriginalName: b.name,
		Description:  b.description,
		ImagePath:    b.imagePath,
	})

	require.NoError(t, err, "failed to create test album")

	// Query to get the album UUID that was just created
	albums, err := db.GetAlbumsForArtist(ctx, sqlhandler.GetAlbumsForArtistParams{
		FromArtist: b.artistUUID,
		Limit:      100,
	})
	require.NoError(t, err, "failed to get albums for artist")
	require.NotEmpty(t, albums, "no albums found for artist")

	// Return the UUID of the most recently created album (last one)
	return albums[len(albums)-1].Uuid
}

// BuildWithData returns the builder data without creating in DB
func (b *AlbumBuilder) BuildWithData() sqlhandler.CreateAlbumParams {
	return sqlhandler.CreateAlbumParams{
		FromArtist:   b.artistUUID,
		OriginalName: b.name,
		Description:  b.description,
		ImagePath:    b.imagePath,
	}
}
