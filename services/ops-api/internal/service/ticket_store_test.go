package service

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestUUIDToString(t *testing.T) {
	tests := []struct {
		name     string
		uuidStr  string
		expected string
		isValid  bool
	}{
		{
			name:     "Valid UUID",
			uuidStr:  "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			expected: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			isValid:  true,
		},
		{
			name:     "Invalid UUID",
			uuidStr:  "",
			expected: "",
			isValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u pgtype.UUID
			if tt.isValid {
				_ = u.Scan(tt.uuidStr)
			} else {
				u.Valid = false
			}

			result := UUIDToString(u)
			if result != tt.expected {
				t.Errorf("UUIDToString() = %v, want %v", result, tt.expected)
			}
		})
	}
}
