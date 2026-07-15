package ideas

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const inboxHeader = "# Idea Inbox\n"

// IdeaEntry represents a single idea extracted from Google Tasks.
type IdeaEntry struct {
	Title       string
	Date        string // YYYY-MM-DD, derived from task.Updated (RFC 3339)
	Account     string // empty in single-account mode
	TaskID      string
	Description string // from task.Notes, may be empty
}

// FormatDate parses an RFC 3339 timestamp and returns YYYY-MM-DD.
func FormatDate(rfc3339 string) string {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return time.Now().Format("2006-01-02")
	}
	return t.Format("2006-01-02")
}

// ReadSyncedTaskIDs reads the inbox file and returns a set of TaskIDs
// already present, parsed from `- TaskID: xxx` lines.
func ReadSyncedTaskIDs(inboxPath string) (map[string]bool, error) {
	ids := make(map[string]bool)

	f, err := os.Open(inboxPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ids, nil
		}
		return nil, fmt.Errorf("opening inbox file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if id, ok := strings.CutPrefix(line, "- TaskID:"); ok {
			id = strings.TrimSpace(id)
			if id != "" {
				ids[id] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading inbox file: %w", err)
	}

	return ids, nil
}

// AppendIdeaEntry appends an idea entry to the inbox file.
// Creates the file with a header if it does not exist.
func AppendIdeaEntry(inboxPath string, entry IdeaEntry) error {
	if err := EnsureInboxFile(inboxPath); err != nil {
		return err
	}

	f, err := os.OpenFile(inboxPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening inbox file for append: %w", err)
	}
	defer f.Close()

	var b strings.Builder
	fmt.Fprintf(&b, "### %s\n", entry.Title)
	fmt.Fprintf(&b, "- Date: %s\n", entry.Date)
	if entry.Account != "" {
		fmt.Fprintf(&b, "- Account: %s\n", entry.Account)
	}
	fmt.Fprintf(&b, "- TaskID: %s\n", entry.TaskID)

	if entry.Description != "" {
		b.WriteString("\n")
		b.WriteString(entry.Description)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	if _, err := f.WriteString(b.String()); err != nil {
		return fmt.Errorf("writing idea entry: %w", err)
	}

	return nil
}

// EnsureInboxFile creates the inbox file with a header if it does not exist.
func EnsureInboxFile(inboxPath string) error {
	if _, err := os.Stat(inboxPath); err == nil {
		return nil
	}

	dir := filepath.Dir(inboxPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating inbox directory: %w", err)
	}

	if err := os.WriteFile(inboxPath, []byte(inboxHeader+"\n"), 0644); err != nil {
		return fmt.Errorf("creating inbox file: %w", err)
	}

	return nil
}
