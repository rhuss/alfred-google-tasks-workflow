package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const accountsFile = "accounts.json"

// validAccountName matches account names: alphanumeric, hyphens, underscores.
var validAccountName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// Account represents a single Google account configuration entry
// within the accounts map in accounts.json.
type Account struct {
	Name         string // symbolic name (e.g., "personal", "work"), set from map key
	Credentials  string `json:"credentials"`   // path to client_secret.json (relative to workflow data dir)
	Keyword      string `json:"keyword"`        // optional Alfred keyword
	ProfileIndex int    `json:"authuser"`       // optional Google multi-login index
}

// AccountConfig represents the parsed accounts.json file that defines
// all configured Google accounts.
type AccountConfig struct {
	Default     string             `json:"default"`      // name of default account
	ListDefault string             `json:"list_default"` // "default" or "all"
	Accounts    map[string]Account `json:"accounts"`     // name -> Account
	dataDir     string             // absolute path to workflow data directory (not serialized)
}

// AccountContext is the runtime-resolved context for a specific account.
// It is created by resolving an account name against the AccountConfig,
// or implicitly for single-account mode (no accounts.json).
type AccountContext struct {
	Name            string // symbolic name of the resolved account; empty for single-account mode
	DataDir         string // absolute path to token storage directory
	CredentialsPath string // absolute path to credentials file
	ProfileIndex    int    // Google authuser index for browser URLs
}

// LoadAccountConfig reads and validates the accounts.json file from the
// given data directory. Returns nil (not an error) if the file does not exist.
func LoadAccountConfig(dataDir string) (*AccountConfig, error) {
	path := filepath.Join(dataDir, accountsFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading accounts config: %w", err)
	}

	var config AccountConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing accounts.json: %w", err)
	}
	config.dataDir = dataDir

	// Normalize account names to lowercase BEFORE validation so that
	// validate() sees the canonical keys (e.g., default lookup works
	// regardless of the case the user typed in accounts.json).
	normalized := make(map[string]Account, len(config.Accounts))
	for name, acct := range config.Accounts {
		lower := strings.ToLower(name)
		if _, exists := normalized[lower]; exists {
			return nil, fmt.Errorf("invalid accounts.json: duplicate account name %q (case-insensitive collision)", name)
		}
		acct.Name = lower
		normalized[lower] = acct
	}
	config.Accounts = normalized

	// Also normalize the default reference
	if config.Default != "" {
		config.Default = strings.ToLower(config.Default)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid accounts.json: %w", err)
	}

	return &config, nil
}

// validate checks the AccountConfig for consistency and correctness.
func (c *AccountConfig) validate() error {
	if len(c.Accounts) == 0 {
		return fmt.Errorf("at least one account must be defined")
	}

	// Validate account names
	for name := range c.Accounts {
		if !validAccountName.MatchString(name) {
			return fmt.Errorf("invalid account name %q: must contain only alphanumeric characters, hyphens, and underscores", name)
		}
	}

	// Validate default account reference
	if c.Default != "" {
		if _, ok := c.Accounts[c.Default]; !ok {
			return fmt.Errorf("default account %q not found in accounts map", c.Default)
		}
	}

	// Validate list_default value
	if c.ListDefault != "" && c.ListDefault != "default" && c.ListDefault != "all" {
		return fmt.Errorf("list_default must be \"default\" or \"all\", got %q", c.ListDefault)
	}

	// Validate credentials paths exist
	for name, acct := range c.Accounts {
		if acct.Credentials == "" {
			return fmt.Errorf("account %q: credentials path is required", name)
		}
		credPath := acct.Credentials
		if !filepath.IsAbs(credPath) {
			credPath = filepath.Join(c.dataDir, credPath)
		}
		if _, err := os.Stat(credPath); os.IsNotExist(err) {
			return fmt.Errorf("account %q: credentials file not found: %s", name, credPath)
		}
	}

	// Validate keyword uniqueness
	keywords := make(map[string]string) // keyword -> account name
	for name, acct := range c.Accounts {
		if acct.Keyword != "" {
			if other, exists := keywords[acct.Keyword]; exists {
				return fmt.Errorf("duplicate keyword %q used by accounts %q and %q", acct.Keyword, other, name)
			}
			keywords[acct.Keyword] = name
		}
	}

	return nil
}

// DefaultContext creates an implicit single-account AccountContext when
// no accounts.json exists. This preserves backward compatibility by
// pointing to the same paths as the current single-account code.
func DefaultContext(dataDir string) *AccountContext {
	return &AccountContext{
		Name:            "",
		DataDir:         dataDir,
		CredentialsPath: filepath.Join(dataDir, credentialsFile),
		ProfileIndex:    -1,
	}
}

// ResolveAccount resolves an account name to an AccountContext using the
// loaded AccountConfig. If name is empty, the default account is used.
func ResolveAccount(config *AccountConfig, name string) (*AccountContext, error) {
	if config == nil {
		return nil, fmt.Errorf("no account configuration loaded")
	}

	// Use default account if no name specified
	if name == "" {
		name = config.defaultAccountName()
	}

	name = strings.ToLower(name)
	acct, ok := config.Accounts[name]
	if !ok {
		return nil, fmt.Errorf("unknown account %q", name)
	}

	// Resolve credentials path
	credPath := acct.Credentials
	if !filepath.IsAbs(credPath) {
		credPath = filepath.Join(config.dataDir, credPath)
	}

	// Token directory is a subdirectory named after the account
	tokenDir := filepath.Join(config.dataDir, name)

	return &AccountContext{
		Name:            name,
		DataDir:         tokenDir,
		CredentialsPath: credPath,
		ProfileIndex:    acct.ProfileIndex,
	}, nil
}

// defaultAccountName returns the name of the default account.
// If Default is set, it is used. Otherwise, the first account
// in alphabetical order is returned.
func (c *AccountConfig) defaultAccountName() string {
	if c.Default != "" {
		return c.Default
	}
	// Return the first account name (map iteration is non-deterministic,
	// so we pick the lexicographically smallest)
	smallest := ""
	for name := range c.Accounts {
		if smallest == "" || name < smallest {
			smallest = name
		}
	}
	return smallest
}

// AccountNames returns a sorted list of all account names.
func (c *AccountConfig) AccountNames() []string {
	names := make([]string, 0, len(c.Accounts))
	for name := range c.Accounts {
		names = append(names, name)
	}
	// Sort for deterministic ordering
	sortStrings(names)
	return names
}

// sortStrings sorts a slice of strings in place (simple insertion sort
// to avoid importing sort for a small slice).
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// FindAccountByKeyword returns the account associated with the given
// Alfred keyword, or nil if no match is found.
func (c *AccountConfig) FindAccountByKeyword(keyword string) *Account {
	for _, acct := range c.Accounts {
		if acct.Keyword == keyword {
			return &acct
		}
	}
	return nil
}

// IsMultiAccount returns true if an AccountConfig is loaded (multi-account mode).
func (c *AccountConfig) IsMultiAccount() bool {
	return c != nil && len(c.Accounts) > 0
}
