package alfred

import (
	"path/filepath"
	"testing"

	aw "github.com/deanishe/awgo"
	taskapi "google.golang.org/api/tasks/v1"

	"github.com/rhuss/alfred-google-tasks-workflow/internal/auth"
	"github.com/rhuss/alfred-google-tasks-workflow/internal/tasks"
)

// newTestWorkflow creates a Workflow suitable for unit testing.
// It sets the required Alfred environment variables and creates
// a temporary data directory.
func newTestWorkflow(t *testing.T) *Workflow {
	t.Helper()
	dir := t.TempDir()

	// aw.New() requires these environment variables
	t.Setenv("alfred_workflow_bundleid", "com.test.workflow")
	t.Setenv("alfred_workflow_cache", filepath.Join(dir, "cache"))
	t.Setenv("alfred_workflow_data", filepath.Join(dir, "data"))

	return &Workflow{
		WF:         aw.New(),
		AccountCtx: auth.DefaultContext(dir),
	}
}

// createTestAccountConfig creates a multi-account AccountConfig for testing.
// It does not need real credential files since we skip validation.
func createTestAccountConfig(dataDir string) *auth.AccountConfig {
	return &auth.AccountConfig{
		Default:     "personal",
		ListDefault: "all",
		Accounts: map[string]auth.Account{
			"personal": {Name: "personal", Credentials: "personal.json"},
			"work":     {Name: "work", Credentials: "work.json", ProfileIndex: 1},
		},
	}
}

func TestExtractAccountPrefix_SingleAccountMode(t *testing.T) {
	w := newTestWorkflow(t)
	// No AccountConfig = single-account mode; @ should be treated as regular text

	tests := []struct {
		input    string
		expected string
	}{
		{"@work add something", "@work add something"},
		{"@personal list", "@personal list"},
		{"list", "list"},
		{"", ""},
		{"@ bare", "@ bare"},
	}

	for _, tt := range tests {
		result, ok := w.extractAccountPrefix(tt.input)
		if !ok {
			t.Errorf("extractAccountPrefix(%q): expected ok=true in single-account mode", tt.input)
		}
		if result != tt.expected {
			t.Errorf("extractAccountPrefix(%q): got %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractAccountPrefix_MultiAccountMode_ValidAccount(t *testing.T) {
	w := newTestWorkflow(t)
	dataDir := w.AccountCtx.DataDir
	w.AccountConfig = createTestAccountConfig(dataDir)

	// Set up a proper ResolveAccount-compatible config with dataDir
	// We need to use a config that ResolveAccount can work with.
	// Since ResolveAccount reads config.dataDir (unexported), we create
	// accounts with absolute credential paths to bypass relative resolution.
	w.AccountConfig = &auth.AccountConfig{
		Default:     "personal",
		ListDefault: "all",
		Accounts: map[string]auth.Account{
			"personal": {Name: "personal", Credentials: filepath.Join(dataDir, "personal.json")},
			"work":     {Name: "work", Credentials: filepath.Join(dataDir, "work.json"), ProfileIndex: 1},
		},
	}

	tests := []struct {
		input         string
		expectedQuery string
		expectedAcct  string
	}{
		{"@work add Buy milk", "add Buy milk", "work"},
		{"@personal list", "list", "personal"},
		{"@work", "", "work"},
		{"@WORK list", "list", "work"}, // case insensitive
	}

	for _, tt := range tests {
		// Reset to default before each test
		w.AccountCtx = auth.DefaultContext(dataDir)
		w.accountTargeted = false

		result, ok := w.extractAccountPrefix(tt.input)
		if !ok {
			t.Errorf("extractAccountPrefix(%q): expected ok=true", tt.input)
			continue
		}
		if result != tt.expectedQuery {
			t.Errorf("extractAccountPrefix(%q): query got %q, want %q", tt.input, result, tt.expectedQuery)
		}
		if w.AccountCtx.Name != tt.expectedAcct {
			t.Errorf("extractAccountPrefix(%q): account got %q, want %q", tt.input, w.AccountCtx.Name, tt.expectedAcct)
		}
	}
}

func TestExtractAccountPrefix_MultiAccountMode_UnknownAccount(t *testing.T) {
	w := newTestWorkflow(t)
	dataDir := w.AccountCtx.DataDir
	w.AccountConfig = &auth.AccountConfig{
		Default: "personal",
		Accounts: map[string]auth.Account{
			"personal": {Name: "personal", Credentials: filepath.Join(dataDir, "personal.json")},
		},
	}

	_, ok := w.extractAccountPrefix("@nonexistent list")
	if ok {
		t.Error("expected ok=false for unknown account name")
	}
}

func TestExtractAccountPrefix_MultiAccountMode_BareAt(t *testing.T) {
	w := newTestWorkflow(t)
	dataDir := w.AccountCtx.DataDir
	w.AccountConfig = &auth.AccountConfig{
		Default: "personal",
		Accounts: map[string]auth.Account{
			"personal": {Name: "personal", Credentials: filepath.Join(dataDir, "personal.json")},
		},
	}

	_, ok := w.extractAccountPrefix("@")
	if ok {
		t.Error("expected ok=false for bare @ (should show account list)")
	}
}

func TestExtractAccountPrefix_MultiAccountMode_NoPrefix(t *testing.T) {
	w := newTestWorkflow(t)
	dataDir := w.AccountCtx.DataDir
	w.AccountConfig = &auth.AccountConfig{
		Default: "personal",
		Accounts: map[string]auth.Account{
			"personal": {Name: "personal", Credentials: filepath.Join(dataDir, "personal.json")},
		},
	}

	result, ok := w.extractAccountPrefix("list")
	if !ok {
		t.Error("expected ok=true for input without @ prefix")
	}
	if result != "list" {
		t.Errorf("expected 'list', got %q", result)
	}
}

func TestBuildSubtitle_SingleAccount(t *testing.T) {
	item := tasks.TaskItem{
		Task:     &taskapi.Task{Title: "Buy milk", Due: "2026-07-01T00:00:00Z"},
		ListName: "Shopping",
		ListID:   "list1",
	}

	subtitle := buildSubtitle("Today", item)
	expected := "[Today] Shopping - due 2026-07-01"
	if subtitle != expected {
		t.Errorf("got %q, want %q", subtitle, expected)
	}
}

func TestBuildSubtitle_MultiAccount(t *testing.T) {
	item := tasks.TaskItem{
		Task:        &taskapi.Task{Title: "Buy milk", Due: "2026-07-01T00:00:00Z"},
		ListName:    "Shopping",
		ListID:      "list1",
		AccountName: "personal",
	}

	subtitle := buildSubtitle("Today", item)
	expected := "[Today] Shopping (personal) - due 2026-07-01"
	if subtitle != expected {
		t.Errorf("got %q, want %q", subtitle, expected)
	}
}

func TestBuildSubtitle_NoDate(t *testing.T) {
	item := tasks.TaskItem{
		Task:        &taskapi.Task{Title: "Someday task"},
		ListName:    "Inbox",
		ListID:      "list1",
		AccountName: "work",
	}

	subtitle := buildSubtitle("No Date", item)
	expected := "[No Date] Inbox (work)"
	if subtitle != expected {
		t.Errorf("got %q, want %q", subtitle, expected)
	}
}

// Verify that in single-account mode, the AccountContext is not changed
func TestExtractAccountPrefix_SingleAccount_PreservesContext(t *testing.T) {
	w := newTestWorkflow(t)
	originalCtx := w.AccountCtx

	_, ok := w.extractAccountPrefix("@anything here")
	if !ok {
		t.Error("expected ok=true in single-account mode")
	}
	if w.AccountCtx != originalCtx {
		t.Error("AccountCtx should not be modified in single-account mode")
	}
}

