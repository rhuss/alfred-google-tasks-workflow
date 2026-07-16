package alfred

import "testing"

func TestEscapeAppleScript(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain", "hello", "hello"},
		{"double quote", `say "hi"`, `say \"hi\"`},
		{"backslash", `a\b`, `a\\b`},
		{"newline", "line1\nline2", `line1\nline2`},
		{"carriage return", "line1\rline2", `line1\rline2`},
		{"tab", "col1\tcol2", `col1\tcol2`},
		{"mixed", "Created: Task\n(due tomorrow)", `Created: Task\n(due tomorrow)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeAppleScript(tt.input); got != tt.expected {
				t.Errorf("escapeAppleScript(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
