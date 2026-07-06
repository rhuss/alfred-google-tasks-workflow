package tasks

import (
	"testing"
	"time"
)

func TestDateToRFC3339(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "simple date",
			input:    time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
			expected: "2026-07-15T00:00:00Z",
		},
		{
			name:     "strips time component",
			input:    time.Date(2026, 7, 15, 14, 30, 45, 0, time.UTC),
			expected: "2026-07-15T00:00:00Z",
		},
		{
			name:     "handles non-UTC timezone",
			input:    time.Date(2026, 7, 15, 23, 0, 0, 0, time.FixedZone("EST", -5*3600)),
			expected: "2026-07-15T00:00:00Z",
		},
		{
			name:     "new year",
			input:    time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2027-01-01T00:00:00Z",
		},
		{
			name:     "leap year date",
			input:    time.Date(2028, 2, 29, 10, 0, 0, 0, time.UTC),
			expected: "2028-02-29T00:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DateToRFC3339(tt.input)
			if result != tt.expected {
				t.Errorf("DateToRFC3339(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
