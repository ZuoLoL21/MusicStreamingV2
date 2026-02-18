package tests

import (
	"file-storage/internal/general"
	"testing"
)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid lowercase", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"valid mixed case", "550e8400-E29B-41d4-A716-446655440000", true},
		{"empty string", "", false},
		{"no dashes", "550e8400e29b41d4a716446655440000", false},
		{"too short", "550e8400-e29b-41d4-a716-44665544000", false},
		{"too long", "550e8400-e29b-41d4-a716-4466554400000", false},
		{"invalid hex chars", "gggggggg-e29b-41d4-a716-446655440000", false},
		{"wrong format", "not-a-valid-uuid-here", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := general.ValidateUUID(tt.input); got != tt.want {
				t.Errorf("ValidateUUID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
