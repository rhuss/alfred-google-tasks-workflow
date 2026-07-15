package ideas

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid RFC3339", "2026-07-13T10:30:00Z", "2026-07-13"},
		{"with timezone offset", "2026-07-13T10:30:00+02:00", "2026-07-13"},
		{"midnight UTC", "2026-01-01T00:00:00Z", "2026-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDate(tt.input)
			if got != tt.expected {
				t.Errorf("FormatDate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatDateInvalid(t *testing.T) {
	got := FormatDate("not-a-date")
	if len(got) != 10 {
		t.Errorf("FormatDate with invalid input should return YYYY-MM-DD format, got %q", got)
	}
}

func TestReadSyncedTaskIDs_ExistingEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inbox.md")

	content := `# Idea Inbox

### First idea
- Date: 2026-07-13
- Account: personal
- TaskID: abc123

### Second idea
- Date: 2026-07-12
- TaskID: def456

Some description here.

`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ids, err := ReadSyncedTaskIDs(path)
	if err != nil {
		t.Fatalf("ReadSyncedTaskIDs() error = %v", err)
	}

	if len(ids) != 2 {
		t.Fatalf("expected 2 task IDs, got %d", len(ids))
	}
	if !ids["abc123"] {
		t.Error("expected abc123 in task IDs")
	}
	if !ids["def456"] {
		t.Error("expected def456 in task IDs")
	}
}

func TestReadSyncedTaskIDs_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inbox.md")

	if err := os.WriteFile(path, []byte("# Idea Inbox\n\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ids, err := ReadSyncedTaskIDs(path)
	if err != nil {
		t.Fatalf("ReadSyncedTaskIDs() error = %v", err)
	}

	if len(ids) != 0 {
		t.Fatalf("expected 0 task IDs, got %d", len(ids))
	}
}

func TestReadSyncedTaskIDs_MissingFile(t *testing.T) {
	ids, err := ReadSyncedTaskIDs("/nonexistent/path/inbox.md")
	if err != nil {
		t.Fatalf("ReadSyncedTaskIDs() error = %v", err)
	}

	if len(ids) != 0 {
		t.Fatalf("expected 0 task IDs for missing file, got %d", len(ids))
	}
}

func TestAppendIdeaEntry_WithDescription(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inbox.md")

	entry := IdeaEntry{
		Title:       "Buy noise-cancelling headphones",
		Date:        "2026-07-13",
		Account:     "personal",
		TaskID:      "dHJhbnNpdC0xMjM",
		Description: "Check reviews on Wirecutter first. Budget around 300 EUR.",
	}

	if err := AppendIdeaEntry(path, entry); err != nil {
		t.Fatalf("AppendIdeaEntry() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	expected := `# Idea Inbox

### Buy noise-cancelling headphones
- Date: 2026-07-13
- Account: personal
- TaskID: dHJhbnNpdC0xMjM

Check reviews on Wirecutter first. Budget around 300 EUR.

`
	if content != expected {
		t.Errorf("file content mismatch.\ngot:\n%s\nwant:\n%s", content, expected)
	}
}

func TestAppendIdeaEntry_WithoutDescription(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inbox.md")

	entry := IdeaEntry{
		Title:   "Research vacation destinations",
		Date:    "2026-07-12",
		Account: "work",
		TaskID:  "dHJhbnNpdC00NTY",
	}

	if err := AppendIdeaEntry(path, entry); err != nil {
		t.Fatalf("AppendIdeaEntry() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	expected := `# Idea Inbox

### Research vacation destinations
- Date: 2026-07-12
- Account: work
- TaskID: dHJhbnNpdC00NTY

`
	if content != expected {
		t.Errorf("file content mismatch.\ngot:\n%s\nwant:\n%s", content, expected)
	}
}

func TestAppendIdeaEntry_WithoutAccount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inbox.md")

	entry := IdeaEntry{
		Title:  "Call dentist",
		Date:   "2026-07-13",
		TaskID: "dHJhbnNpdC03ODk",
	}

	if err := AppendIdeaEntry(path, entry); err != nil {
		t.Fatalf("AppendIdeaEntry() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	expected := `# Idea Inbox

### Call dentist
- Date: 2026-07-13
- TaskID: dHJhbnNpdC03ODk

`
	if content != expected {
		t.Errorf("file content mismatch.\ngot:\n%s\nwant:\n%s", content, expected)
	}
}

func TestAppendIdeaEntry_PreservesExistingContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inbox.md")

	existing := `# Idea Inbox

### Existing idea
- Date: 2026-07-10
- TaskID: existing123

`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	entry := IdeaEntry{
		Title:  "New idea",
		Date:   "2026-07-13",
		TaskID: "new456",
	}

	if err := AppendIdeaEntry(path, entry); err != nil {
		t.Fatalf("AppendIdeaEntry() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	expected := `# Idea Inbox

### Existing idea
- Date: 2026-07-10
- TaskID: existing123

### New idea
- Date: 2026-07-13
- TaskID: new456

`
	if content != expected {
		t.Errorf("file content mismatch.\ngot:\n%s\nwant:\n%s", content, expected)
	}
}

func TestAppendIdeaEntry_CreatesParentDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "nested", "dir", "inbox.md")

	entry := IdeaEntry{
		Title:  "Nested idea",
		Date:   "2026-07-13",
		TaskID: "nested789",
	}

	if err := AppendIdeaEntry(path, entry); err != nil {
		t.Fatalf("AppendIdeaEntry() error = %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected inbox file to be created")
	}

	ids, err := ReadSyncedTaskIDs(path)
	if err != nil {
		t.Fatal(err)
	}
	if !ids["nested789"] {
		t.Error("expected nested789 in task IDs after creation")
	}
}

func TestAppendIdeaEntry_DeeplyNestedParentCreation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "d", "e", "inbox.md")

	entry := IdeaEntry{
		Title:  "Deep idea",
		Date:   "2026-07-13",
		TaskID: "deep001",
	}

	if err := AppendIdeaEntry(path, entry); err != nil {
		t.Fatalf("AppendIdeaEntry() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if content[:len(inboxHeader)] != inboxHeader {
		t.Errorf("expected file to start with header %q", inboxHeader)
	}

	ids, err := ReadSyncedTaskIDs(path)
	if err != nil {
		t.Fatal(err)
	}
	if !ids["deep001"] {
		t.Error("expected deep001 in task IDs")
	}
}

func TestAppendIdeaEntry_SubsequentAppendsPreserveHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inbox.md")

	entries := []IdeaEntry{
		{Title: "First", Date: "2026-07-11", TaskID: "first001"},
		{Title: "Second", Date: "2026-07-12", TaskID: "second002"},
		{Title: "Third", Date: "2026-07-13", TaskID: "third003"},
	}

	for _, entry := range entries {
		if err := AppendIdeaEntry(path, entry); err != nil {
			t.Fatalf("AppendIdeaEntry(%q) error = %v", entry.Title, err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	if content[:len(inboxHeader)] != inboxHeader {
		t.Error("header should be preserved after multiple appends")
	}

	headerCount := 0
	for i := 0; i < len(content)-len("# Idea Inbox"); i++ {
		if content[i:i+len("# Idea Inbox")] == "# Idea Inbox" {
			headerCount++
		}
	}
	if headerCount != 1 {
		t.Errorf("expected exactly 1 header, found %d", headerCount)
	}

	ids, err := ReadSyncedTaskIDs(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 {
		t.Errorf("expected 3 task IDs, got %d", len(ids))
	}
}

func TestAppendIdeaEntry_AutoCreateWithHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new-inbox.md")

	entry := IdeaEntry{
		Title:  "First idea",
		Date:   "2026-07-13",
		TaskID: "first001",
	}

	if err := AppendIdeaEntry(path, entry); err != nil {
		t.Fatalf("AppendIdeaEntry() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if content[:len(inboxHeader)] != inboxHeader {
		t.Errorf("expected file to start with header %q, got %q", inboxHeader, content[:20])
	}
}
