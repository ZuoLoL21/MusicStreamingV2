//go:build integration

package builders

import "github.com/jackc/pgx/v5/pgtype"

// NewPgtypeText creates a pgtype.Text from a string
// Returns Valid=false if string is empty
func NewPgtypeText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// NewPgtypeBool creates a pgtype.Bool from a bool
func NewPgtypeBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

// NewPgtypeInt4 creates a pgtype.Int4 from an int32
func NewPgtypeInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}
