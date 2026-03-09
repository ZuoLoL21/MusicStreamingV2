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

// PlaylistBuilder provides a fluent interface for creating test playlists
type PlaylistBuilder struct {
	name        string
	description pgtype.Text
	isPublic    pgtype.Bool
	imagePath   pgtype.Text
	userUUID    pgtype.UUID
}

// NewPlaylistBuilder creates a new PlaylistBuilder with random defaults
func NewPlaylistBuilder(userUUID pgtype.UUID) *PlaylistBuilder {
	id := uuid.New().String()[:8]
	return &PlaylistBuilder{
		name:        fmt.Sprintf("Playlist_%s", id),
		description: NewPgtypeText("Test playlist description"),
		isPublic:    NewPgtypeBool(true),
		imagePath:   NewPgtypeText(""),
		userUUID:    userUUID,
	}
}

// WithName sets the playlist name
func (b *PlaylistBuilder) WithName(name string) *PlaylistBuilder {
	b.name = name
	return b
}

// WithDescription sets the description
func (b *PlaylistBuilder) WithDescription(description string) *PlaylistBuilder {
	b.description = NewPgtypeText(description)
	return b
}

// WithPublic sets whether the playlist is public
func (b *PlaylistBuilder) WithPublic(isPublic bool) *PlaylistBuilder {
	b.isPublic = NewPgtypeBool(isPublic)
	return b
}

// WithImagePath sets the image path
func (b *PlaylistBuilder) WithImagePath(imagePath string) *PlaylistBuilder {
	b.imagePath = NewPgtypeText(imagePath)
	return b
}

// Build creates the playlist in the database and returns the playlist UUID
func (b *PlaylistBuilder) Build(t *testing.T, ctx context.Context, db interface {
	CreatePlaylist(ctx context.Context, arg sqlhandler.CreatePlaylistParams) error
	GetPlaylistsForUser(ctx context.Context, arg sqlhandler.GetPlaylistsForUserParams) ([]sqlhandler.Playlist, error)
}) pgtype.UUID {
	t.Helper()

	err := db.CreatePlaylist(ctx, sqlhandler.CreatePlaylistParams{
		FromUser:     b.userUUID,
		OriginalName: b.name,
		Description:  b.description,
		IsPublic:     b.isPublic,
		ImagePath:    b.imagePath,
	})

	require.NoError(t, err, "failed to create test playlist")

	// Query to get the playlist UUID that was just created
	playlists, err := db.GetPlaylistsForUser(ctx, sqlhandler.GetPlaylistsForUserParams{
		FromUser: b.userUUID,
		Limit:    100,
	})
	require.NoError(t, err, "failed to get playlists for user")
	require.NotEmpty(t, playlists, "no playlists found for user")

	// Return the UUID of the most recently created playlist (last one)
	return playlists[len(playlists)-1].Uuid
}

// BuildWithData returns the builder data without creating in DB
func (b *PlaylistBuilder) BuildWithData() sqlhandler.CreatePlaylistParams {
	return sqlhandler.CreatePlaylistParams{
		FromUser:     b.userUUID,
		OriginalName: b.name,
		Description:  b.description,
		IsPublic:     b.isPublic,
		ImagePath:    b.imagePath,
	}
}
