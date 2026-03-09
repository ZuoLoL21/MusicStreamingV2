//go:build integration

package builders

import (
	sqlhandler "backend/sql/sqlc"
	"context"
	"fmt"
	"libs/helpers"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

// UserBuilder provides a fluent interface for creating test users
type UserBuilder struct {
	username         string
	email            string
	hashedPassword   string
	bio              string
	profileImagePath string
	country          string
}

// NewUserBuilder creates a new UserBuilder with random defaults
func NewUserBuilder() *UserBuilder {
	id := uuid.New().String()[:8]
	return &UserBuilder{
		username:         fmt.Sprintf("user_%s", id),
		email:            fmt.Sprintf("user_%s@test.com", id),
		hashedPassword:   "Test123!@#", // Will be hashed in Build()
		bio:              "Test user bio",
		profileImagePath: "",
		country:          "US",
	}
}

// WithUsername sets the username
func (b *UserBuilder) WithUsername(username string) *UserBuilder {
	b.username = username
	return b
}

// WithEmail sets the email
func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.email = email
	return b
}

// WithPassword sets the password (will be hashed in Build())
func (b *UserBuilder) WithPassword(password string) *UserBuilder {
	b.hashedPassword = password
	return b
}

// WithHashedPassword sets the hashed password directly
func (b *UserBuilder) WithHashedPassword(hashedPassword string) *UserBuilder {
	b.hashedPassword = hashedPassword
	return b
}

// WithBio sets the bio
func (b *UserBuilder) WithBio(bio string) *UserBuilder {
	b.bio = bio
	return b
}

// WithProfileImagePath sets the profile image path
func (b *UserBuilder) WithProfileImagePath(profileImagePath string) *UserBuilder {
	b.profileImagePath = profileImagePath
	return b
}

// WithCountry sets the country
func (b *UserBuilder) WithCountry(country string) *UserBuilder {
	b.country = country
	return b
}

// Build creates the user in the database and returns the UUID
func (b *UserBuilder) Build(t *testing.T, ctx context.Context, db interface {
	CreateUser(ctx context.Context, arg sqlhandler.CreateUserParams) (pgtype.UUID, error)
}) pgtype.UUID {
	t.Helper()

	// Hash the password
	hashedPassword := helpers.Encode(b.hashedPassword)

	userUUID, err := db.CreateUser(ctx, sqlhandler.CreateUserParams{
		Username:         b.username,
		Email:            b.email,
		HashedPassword:   hashedPassword,
		Bio:              NewPgtypeText(b.bio),
		ProfileImagePath: NewPgtypeText(b.profileImagePath),
		Country:          b.country,
	})

	require.NoError(t, err, "failed to create test user")
	require.True(t, userUUID.Valid, "user UUID should be valid")

	return userUUID
}

// BuildWithData returns the builder data without creating in DB
func (b *UserBuilder) BuildWithData() sqlhandler.CreateUserParams {
	hashedPassword := helpers.Encode(b.hashedPassword)

	return sqlhandler.CreateUserParams{
		Username:         b.username,
		Email:            b.email,
		HashedPassword:   hashedPassword,
		Bio:              NewPgtypeText(b.bio),
		ProfileImagePath: NewPgtypeText(b.profileImagePath),
		Country:          b.country,
	}
}
