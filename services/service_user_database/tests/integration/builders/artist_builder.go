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

// ArtistBuilder provides a fluent interface for creating test artists
type ArtistBuilder struct {
	userUUID         pgtype.UUID
	name             string
	bio              string
	profileImagePath string
}

// NewArtistBuilder creates a new ArtistBuilder with random defaults
// It requires a userUUID who will become the owner of the artist
func NewArtistBuilder(userUUID pgtype.UUID) *ArtistBuilder {
	id := uuid.New().String()[:8]
	return &ArtistBuilder{
		userUUID:         userUUID,
		name:             fmt.Sprintf("Artist_%s", id),
		bio:              "Test artist bio",
		profileImagePath: "",
	}
}

// WithName sets the artist name
func (b *ArtistBuilder) WithName(name string) *ArtistBuilder {
	b.name = name
	return b
}

// WithBio sets the bio
func (b *ArtistBuilder) WithBio(bio string) *ArtistBuilder {
	b.bio = bio
	return b
}

// WithProfileImagePath sets the profile image path
func (b *ArtistBuilder) WithProfileImagePath(profileImagePath string) *ArtistBuilder {
	b.profileImagePath = profileImagePath
	return b
}

// Build creates the artist in the database and returns the artist UUID
func (b *ArtistBuilder) Build(t *testing.T, ctx context.Context, db interface {
	CreateArtist(ctx context.Context, arg sqlhandler.CreateArtistParams) error
	GetArtistForUser(ctx context.Context, userUuid pgtype.UUID) ([]sqlhandler.GetArtistForUserRow, error)
}) pgtype.UUID {
	t.Helper()

	err := db.CreateArtist(ctx, sqlhandler.CreateArtistParams{
		UserUuid:         b.userUUID,
		ArtistName:       b.name,
		Bio:              NewPgtypeText(b.bio),
		ProfileImagePath: NewPgtypeText(b.profileImagePath),
	})

	require.NoError(t, err, "failed to create test artist")

	// Query to get the artist UUID that was just created
	artists, err := db.GetArtistForUser(ctx, b.userUUID)
	require.NoError(t, err, "failed to get artist for user")
	require.NotEmpty(t, artists, "no artists found for user")

	// Return the UUID of the most recently created artist (last one)
	return artists[len(artists)-1].Uuid
}

// BuildWithData returns the builder data without creating in DB
func (b *ArtistBuilder) BuildWithData() sqlhandler.CreateArtistParams {
	return sqlhandler.CreateArtistParams{
		UserUuid:         b.userUUID,
		ArtistName:       b.name,
		Bio:              NewPgtypeText(b.bio),
		ProfileImagePath: NewPgtypeText(b.profileImagePath),
	}
}
