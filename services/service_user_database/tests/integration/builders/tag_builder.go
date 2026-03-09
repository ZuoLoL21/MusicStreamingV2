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

// TagBuilder provides a fluent interface for creating test tags
type TagBuilder struct {
	name        string
	description pgtype.Text
}

// NewTagBuilder creates a new TagBuilder with random defaults
func NewTagBuilder() *TagBuilder {
	id := uuid.New().String()[:8]
	return &TagBuilder{
		name:        fmt.Sprintf("tag_%s", id),
		description: pgtype.Text{String: "Test tag description", Valid: true},
	}
}

// WithName sets the tag name
func (b *TagBuilder) WithName(name string) *TagBuilder {
	b.name = name
	return b
}

// WithDescription sets the description
func (b *TagBuilder) WithDescription(description string) *TagBuilder {
	b.description = pgtype.Text{String: description, Valid: true}
	return b
}

// Build creates the tag in the database
func (b *TagBuilder) Build(t *testing.T, ctx context.Context, db interface {
	CreateTag(ctx context.Context, arg sqlhandler.CreateTagParams) error
}) {
	t.Helper()

	err := db.CreateTag(ctx, sqlhandler.CreateTagParams{
		TagName:        b.name,
		TagDescription: b.description,
	})

	require.NoError(t, err, "failed to create test tag")
}

// BuildWithData returns the builder data without creating in DB
func (b *TagBuilder) BuildWithData() sqlhandler.CreateTagParams {
	return sqlhandler.CreateTagParams{
		TagName:        b.name,
		TagDescription: b.description,
	}
}
