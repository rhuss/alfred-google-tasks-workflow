package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	credentialsFile = "client_secret.json"
	// Google Tasks read-write scope
	TasksScope = "https://www.googleapis.com/auth/tasks"
)

// LoadClientCredentials reads and parses the client_secret.json file from the
// given data directory. Returns an oauth2.Config configured for the desktop
// OAuth flow with the Google Tasks scope.
func LoadClientCredentials(dataDir string) (*oauth2.Config, error) {
	path := filepath.Join(dataDir, credentialsFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("credentials file not found: %s\nDownload client_secret.json from Google Cloud Console and place it in:\n  %s", path, dataDir)
		}
		return nil, fmt.Errorf("reading credentials file: %w", err)
	}

	config, err := google.ConfigFromJSON(data, TasksScope)
	if err != nil {
		// Try parsing manually for "installed" type credentials
		var raw struct {
			Installed *struct {
				ClientID     string   `json:"client_id"`
				ClientSecret string   `json:"client_secret"`
				AuthURI      string   `json:"auth_uri"`
				TokenURI     string   `json:"token_uri"`
				RedirectURIs []string `json:"redirect_uris"`
			} `json:"installed"`
		}
		if jsonErr := json.Unmarshal(data, &raw); jsonErr != nil || raw.Installed == nil {
			return nil, fmt.Errorf("invalid credentials file format: %w", err)
		}

		config = &oauth2.Config{
			ClientID:     raw.Installed.ClientID,
			ClientSecret: raw.Installed.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  raw.Installed.AuthURI,
				TokenURL: raw.Installed.TokenURI,
			},
			RedirectURL: "http://127.0.0.1",
			Scopes:      []string{TasksScope},
		}
	}

	return config, nil
}
