package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestDir creates a temporary directory with optional accounts.json content
// and credential files.
func createTestDir(t *testing.T, accountsJSON string, credFiles []string) string {
	t.Helper()
	dir := t.TempDir()

	if accountsJSON != "" {
		if err := os.WriteFile(filepath.Join(dir, "accounts.json"), []byte(accountsJSON), 0644); err != nil {
			t.Fatalf("writing accounts.json: %v", err)
		}
	}

	for _, cf := range credFiles {
		path := filepath.Join(dir, cf)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("creating dir for %s: %v", cf, err)
		}
		// Write a minimal valid credentials file
		if err := os.WriteFile(path, []byte(`{"installed":{"client_id":"test","client_secret":"test"}}`), 0644); err != nil {
			t.Fatalf("writing %s: %v", cf, err)
		}
	}

	return dir
}

func TestLoadAccountConfig_NoFile(t *testing.T) {
	dir := t.TempDir()
	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config != nil {
		t.Fatal("expected nil config when no accounts.json exists")
	}
}

func TestLoadAccountConfig_ValidConfig(t *testing.T) {
	json := `{
		"default": "personal",
		"list_default": "all",
		"accounts": {
			"personal": {
				"credentials": "client_secret.json"
			},
			"work": {
				"credentials": "work/client_secret.json",
				"authuser": 1
			}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json", "work/client_secret.json"})

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Default != "personal" {
		t.Errorf("expected default 'personal', got %q", config.Default)
	}
	if config.ListDefault != "all" {
		t.Errorf("expected list_default 'all', got %q", config.ListDefault)
	}
	if len(config.Accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(config.Accounts))
	}
	if config.Accounts["work"].ProfileIndex != 1 {
		t.Errorf("expected work authuser 1, got %d", config.Accounts["work"].ProfileIndex)
	}
	// Verify Name field is populated from map key
	if config.Accounts["personal"].Name != "personal" {
		t.Errorf("expected personal account Name 'personal', got %q", config.Accounts["personal"].Name)
	}
	if config.Accounts["work"].Name != "work" {
		t.Errorf("expected work account Name 'work', got %q", config.Accounts["work"].Name)
	}
}

func TestLoadAccountConfig_InvalidJSON(t *testing.T) {
	dir := createTestDir(t, `{not valid json}`, nil)

	_, err := LoadAccountConfig(dir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadAccountConfig_EmptyAccounts(t *testing.T) {
	json := `{"accounts": {}}`
	dir := createTestDir(t, json, nil)

	_, err := LoadAccountConfig(dir)
	if err == nil {
		t.Fatal("expected error for empty accounts")
	}
}

func TestLoadAccountConfig_InvalidDefaultRef(t *testing.T) {
	json := `{
		"default": "nonexistent",
		"accounts": {
			"personal": {"credentials": "client_secret.json"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json"})

	_, err := LoadAccountConfig(dir)
	if err == nil {
		t.Fatal("expected error for invalid default reference")
	}
}

func TestLoadAccountConfig_InvalidListDefault(t *testing.T) {
	json := `{
		"list_default": "bogus",
		"accounts": {
			"personal": {"credentials": "client_secret.json"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json"})

	_, err := LoadAccountConfig(dir)
	if err == nil {
		t.Fatal("expected error for invalid list_default")
	}
}

func TestLoadAccountConfig_MissingCredentials(t *testing.T) {
	json := `{
		"accounts": {
			"personal": {"credentials": "nonexistent.json"}
		}
	}`
	dir := createTestDir(t, json, nil)

	_, err := LoadAccountConfig(dir)
	if err == nil {
		t.Fatal("expected error for missing credentials file")
	}
}

func TestLoadAccountConfig_EmptyCredentialsPath(t *testing.T) {
	json := `{
		"accounts": {
			"personal": {"credentials": ""}
		}
	}`
	dir := createTestDir(t, json, nil)

	_, err := LoadAccountConfig(dir)
	if err == nil {
		t.Fatal("expected error for empty credentials path")
	}
}

func TestLoadAccountConfig_InvalidAccountName(t *testing.T) {
	json := `{
		"accounts": {
			"my account": {"credentials": "client_secret.json"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json"})

	_, err := LoadAccountConfig(dir)
	if err == nil {
		t.Fatal("expected error for invalid account name with spaces")
	}
}

func TestLoadAccountConfig_DuplicateKeywords(t *testing.T) {
	json := `{
		"accounts": {
			"personal": {"credentials": "client_secret.json", "keyword": "gt"},
			"work": {"credentials": "work/client_secret.json", "keyword": "gt"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json", "work/client_secret.json"})

	_, err := LoadAccountConfig(dir)
	if err == nil {
		t.Fatal("expected error for duplicate keywords")
	}
}

func TestDefaultContext(t *testing.T) {
	dir := "/test/data"
	ctx := DefaultContext(dir)

	if ctx.Name != "" {
		t.Errorf("expected empty name, got %q", ctx.Name)
	}
	if ctx.DataDir != dir {
		t.Errorf("expected DataDir %q, got %q", dir, ctx.DataDir)
	}
	expectedCred := filepath.Join(dir, "client_secret.json")
	if ctx.CredentialsPath != expectedCred {
		t.Errorf("expected CredentialsPath %q, got %q", expectedCred, ctx.CredentialsPath)
	}
	if ctx.ProfileIndex != 0 {
		t.Errorf("expected ProfileIndex 0, got %d", ctx.ProfileIndex)
	}
}

func TestResolveAccount_Default(t *testing.T) {
	json := `{
		"default": "work",
		"accounts": {
			"personal": {"credentials": "client_secret.json"},
			"work": {"credentials": "work/client_secret.json", "authuser": 2}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json", "work/client_secret.json"})

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Resolve with empty name should use default
	ctx, err := ResolveAccount(config, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Name != "work" {
		t.Errorf("expected name 'work', got %q", ctx.Name)
	}
	expectedDataDir := filepath.Join(dir, "work")
	if ctx.DataDir != expectedDataDir {
		t.Errorf("expected DataDir %q, got %q", expectedDataDir, ctx.DataDir)
	}
	expectedCred := filepath.Join(dir, "work/client_secret.json")
	if ctx.CredentialsPath != expectedCred {
		t.Errorf("expected CredentialsPath %q, got %q", expectedCred, ctx.CredentialsPath)
	}
	if ctx.ProfileIndex != 2 {
		t.Errorf("expected ProfileIndex 2, got %d", ctx.ProfileIndex)
	}
}

func TestResolveAccount_ByName(t *testing.T) {
	json := `{
		"default": "work",
		"accounts": {
			"personal": {"credentials": "client_secret.json"},
			"work": {"credentials": "work/client_secret.json"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json", "work/client_secret.json"})

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, err := ResolveAccount(config, "personal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Name != "personal" {
		t.Errorf("expected name 'personal', got %q", ctx.Name)
	}
}

func TestResolveAccount_UnknownName(t *testing.T) {
	json := `{
		"accounts": {
			"personal": {"credentials": "client_secret.json"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json"})

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = ResolveAccount(config, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown account name")
	}
}

func TestResolveAccount_NilConfig(t *testing.T) {
	_, err := ResolveAccount(nil, "anything")
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestAccountConfig_DefaultAccountName_Implicit(t *testing.T) {
	json := `{
		"accounts": {
			"beta": {"credentials": "client_secret.json"},
			"alpha": {"credentials": "client_secret.json"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json"})

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Without explicit default, should pick lexicographically smallest
	ctx, err := ResolveAccount(config, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Name != "alpha" {
		t.Errorf("expected implicit default 'alpha', got %q", ctx.Name)
	}
}

func TestAccountConfig_AccountNames(t *testing.T) {
	json := `{
		"accounts": {
			"charlie": {"credentials": "client_secret.json"},
			"alpha": {"credentials": "client_secret.json"},
			"bravo": {"credentials": "client_secret.json"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json"})

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := config.AccountNames()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	if names[0] != "alpha" || names[1] != "bravo" || names[2] != "charlie" {
		t.Errorf("expected sorted names [alpha bravo charlie], got %v", names)
	}
}

func TestAccountConfig_FindAccountByKeyword(t *testing.T) {
	json := `{
		"accounts": {
			"personal": {"credentials": "client_secret.json"},
			"work": {"credentials": "work/client_secret.json", "keyword": "gtw"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json", "work/client_secret.json"})

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	acct := config.FindAccountByKeyword("gtw")
	if acct == nil {
		t.Fatal("expected to find account with keyword 'gtw'")
	}
	if acct.Name != "work" {
		t.Errorf("expected account 'work', got %q", acct.Name)
	}

	acct = config.FindAccountByKeyword("nonexistent")
	if acct != nil {
		t.Fatal("expected nil for nonexistent keyword")
	}
}

func TestAccountConfig_IsMultiAccount(t *testing.T) {
	var nilConfig *AccountConfig
	if nilConfig.IsMultiAccount() {
		t.Error("nil config should not be multi-account")
	}

	json := `{
		"accounts": {
			"personal": {"credentials": "client_secret.json"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json"})

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !config.IsMultiAccount() {
		t.Error("loaded config should be multi-account")
	}
}

func TestResolveAccount_CaseInsensitive(t *testing.T) {
	json := `{
		"accounts": {
			"work": {"credentials": "client_secret.json"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json"})

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, err := ResolveAccount(config, "Work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Name != "work" {
		t.Errorf("expected name 'work', got %q", ctx.Name)
	}
}

func TestLoadAccountConfig_ListDefaultDefaults(t *testing.T) {
	json := `{
		"accounts": {
			"personal": {"credentials": "client_secret.json"}
		}
	}`
	dir := createTestDir(t, json, []string{"client_secret.json"})

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// list_default omitted should be empty string (caller defaults to "default")
	if config.ListDefault != "" {
		t.Errorf("expected empty list_default, got %q", config.ListDefault)
	}
}

func TestLoadAccountConfig_ValidListDefaults(t *testing.T) {
	for _, ld := range []string{"default", "all"} {
		json := `{
			"list_default": "` + ld + `",
			"accounts": {
				"personal": {"credentials": "client_secret.json"}
			}
		}`
		dir := createTestDir(t, json, []string{"client_secret.json"})

		config, err := LoadAccountConfig(dir)
		if err != nil {
			t.Fatalf("unexpected error for list_default=%q: %v", ld, err)
		}
		if config.ListDefault != ld {
			t.Errorf("expected list_default %q, got %q", ld, config.ListDefault)
		}
	}
}

// --- Integration tests: full load-resolve-context round-trip ---

func TestIntegration_FullMultiAccountRoundTrip(t *testing.T) {
	// Set up a multi-account config with two accounts, each with its own credentials
	configJSON := `{
		"default": "work",
		"list_default": "all",
		"accounts": {
			"personal": {
				"credentials": "personal_secret.json",
				"keyword": "gtp"
			},
			"work": {
				"credentials": "work/client_secret.json",
				"authuser": 1,
				"keyword": "gtw"
			}
		}
	}`
	dir := createTestDir(t, configJSON, []string{
		"personal_secret.json",
		"work/client_secret.json",
	})

	// Step 1: Load config
	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("LoadAccountConfig failed: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil config")
	}

	// Step 2: Verify multi-account mode
	if !config.IsMultiAccount() {
		t.Error("expected IsMultiAccount() to be true")
	}

	// Step 3: Resolve default account (should be "work")
	ctx, err := ResolveAccount(config, "")
	if err != nil {
		t.Fatalf("ResolveAccount default failed: %v", err)
	}
	if ctx.Name != "work" {
		t.Errorf("expected default account 'work', got %q", ctx.Name)
	}
	if ctx.ProfileIndex != 1 {
		t.Errorf("expected ProfileIndex 1, got %d", ctx.ProfileIndex)
	}
	expectedTokenDir := filepath.Join(dir, "work")
	if ctx.DataDir != expectedTokenDir {
		t.Errorf("expected DataDir %q, got %q", expectedTokenDir, ctx.DataDir)
	}
	expectedCred := filepath.Join(dir, "work/client_secret.json")
	if ctx.CredentialsPath != expectedCred {
		t.Errorf("expected CredentialsPath %q, got %q", expectedCred, ctx.CredentialsPath)
	}

	// Step 4: Resolve explicit account "personal"
	ctx, err = ResolveAccount(config, "personal")
	if err != nil {
		t.Fatalf("ResolveAccount personal failed: %v", err)
	}
	if ctx.Name != "personal" {
		t.Errorf("expected name 'personal', got %q", ctx.Name)
	}
	if ctx.ProfileIndex != 0 {
		t.Errorf("expected ProfileIndex 0, got %d", ctx.ProfileIndex)
	}
	expectedCred = filepath.Join(dir, "personal_secret.json")
	if ctx.CredentialsPath != expectedCred {
		t.Errorf("expected CredentialsPath %q, got %q", expectedCred, ctx.CredentialsPath)
	}

	// Step 5: Verify AccountNames returns sorted list
	names := config.AccountNames()
	if len(names) != 2 || names[0] != "personal" || names[1] != "work" {
		t.Errorf("expected [personal work], got %v", names)
	}

	// Step 6: Verify keyword lookup
	acct := config.FindAccountByKeyword("gtw")
	if acct == nil || acct.Name != "work" {
		t.Error("expected FindAccountByKeyword('gtw') to return work account")
	}
	acct = config.FindAccountByKeyword("gtp")
	if acct == nil || acct.Name != "personal" {
		t.Error("expected FindAccountByKeyword('gtp') to return personal account")
	}
}

func TestIntegration_SingleAccountFallback(t *testing.T) {
	// No accounts.json: single-account mode
	dir := t.TempDir()

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config != nil {
		t.Fatal("expected nil config for single-account mode")
	}

	// DefaultContext should work as fallback
	ctx := DefaultContext(dir)
	if ctx.Name != "" {
		t.Errorf("expected empty name for single-account, got %q", ctx.Name)
	}
	if ctx.DataDir != dir {
		t.Errorf("expected DataDir %q, got %q", dir, ctx.DataDir)
	}
	expectedCred := filepath.Join(dir, "client_secret.json")
	if ctx.CredentialsPath != expectedCred {
		t.Errorf("expected CredentialsPath %q, got %q", expectedCred, ctx.CredentialsPath)
	}
}

func TestIntegration_InvalidConfig_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		credFiles  []string
		errSubstr  string
	}{
		{
			name:      "malformed JSON",
			json:      `{not valid}`,
			errSubstr: "parsing accounts.json",
		},
		{
			name:      "empty accounts map",
			json:      `{"accounts": {}}`,
			errSubstr: "at least one account",
		},
		{
			name:      "bad default reference",
			json:      `{"default": "ghost", "accounts": {"work": {"credentials": "c.json"}}}`,
			credFiles: []string{"c.json"},
			errSubstr: "default account",
		},
		{
			name:      "invalid list_default value",
			json:      `{"list_default": "bogus", "accounts": {"work": {"credentials": "c.json"}}}`,
			credFiles: []string{"c.json"},
			errSubstr: "list_default",
		},
		{
			name:      "missing credentials file on disk",
			json:      `{"accounts": {"work": {"credentials": "missing.json"}}}`,
			errSubstr: "credentials file not found",
		},
		{
			name:      "empty credentials path",
			json:      `{"accounts": {"work": {"credentials": ""}}}`,
			errSubstr: "credentials path is required",
		},
		{
			name:      "invalid account name",
			json:      `{"accounts": {"my account": {"credentials": "c.json"}}}`,
			credFiles: []string{"c.json"},
			errSubstr: "invalid account name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := createTestDir(t, tt.json, tt.credFiles)
			config, err := LoadAccountConfig(dir)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if config != nil {
				t.Fatal("expected nil config on error")
			}
			if !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("expected error to contain %q, got: %v", tt.errSubstr, err)
			}
		})
	}
}

func TestLoadAccountConfig_CaseCollision(t *testing.T) {
	// Two account names that differ only in case should produce an error
	configJSON := `{
		"accounts": {
			"Work": {"credentials": "c.json"},
			"work": {"credentials": "c.json"}
		}
	}`
	dir := createTestDir(t, configJSON, []string{"c.json"})

	_, err := LoadAccountConfig(dir)
	if err == nil {
		t.Fatal("expected error for case-insensitive duplicate account names")
	}
	if !strings.Contains(err.Error(), "duplicate account name") {
		t.Errorf("expected 'duplicate account name' error, got: %v", err)
	}
}

func TestIntegration_CaseNormalization(t *testing.T) {
	// Account names with mixed case in JSON should be normalized to lowercase
	configJSON := `{
		"default": "Work",
		"accounts": {
			"Work": {"credentials": "c.json"},
			"Personal": {"credentials": "c.json"}
		}
	}`
	dir := createTestDir(t, configJSON, []string{"c.json"})

	config, err := LoadAccountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Keys should be lowercase
	if _, ok := config.Accounts["work"]; !ok {
		t.Error("expected lowercase key 'work' in accounts map")
	}
	if _, ok := config.Accounts["personal"]; !ok {
		t.Error("expected lowercase key 'personal' in accounts map")
	}
	if _, ok := config.Accounts["Work"]; ok {
		t.Error("original case key 'Work' should not exist after normalization")
	}

	// Default should be lowercased
	if config.Default != "work" {
		t.Errorf("expected default 'work', got %q", config.Default)
	}

	// Resolve with any case should work
	ctx, err := ResolveAccount(config, "WORK")
	if err != nil {
		t.Fatalf("ResolveAccount WORK failed: %v", err)
	}
	if ctx.Name != "work" {
		t.Errorf("expected resolved name 'work', got %q", ctx.Name)
	}
}
