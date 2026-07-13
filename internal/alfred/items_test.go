package alfred

import "testing"

func TestLargetypeText(t *testing.T) {
	tests := []struct {
		name  string
		notes string
		want  string
	}{
		{"non-empty notes", "Buy milk and eggs", "Buy milk and eggs"},
		{"empty string", "", "(no details)"},
		{"whitespace only spaces", "   ", "(no details)"},
		{"whitespace only tabs", "\t\t", "(no details)"},
		{"whitespace only mixed", "  \t\n  ", "(no details)"},
		{"notes with leading whitespace", "  some notes", "  some notes"},
		{"unicode content", "Task with emoji 🎉", "Task with emoji 🎉"},
		{"multiline notes", "Line 1\nLine 2\nLine 3", "Line 1\nLine 2\nLine 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := largetypeText(tt.notes)
			if got != tt.want {
				t.Errorf("largetypeText(%q) = %q, want %q", tt.notes, got, tt.want)
			}
		})
	}
}
