package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

const tokenFile = "token.json"

// LoadToken reads the stored OAuth token from the data directory.
func LoadToken(dataDir string) (*oauth2.Token, error) {
	path := filepath.Join(dataDir, tokenFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not authenticated: run 'gt login' first")
		}
		return nil, fmt.Errorf("reading token file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing token file: %w", err)
	}

	return &token, nil
}

// SaveToken writes the OAuth token to the data directory.
func SaveToken(dataDir string, token *oauth2.Token) error {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	path := filepath.Join(dataDir, tokenFile)
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding token: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing token file: %w", err)
	}

	return nil
}

// DeleteToken removes the stored OAuth token from the data directory.
func DeleteToken(dataDir string) error {
	path := filepath.Join(dataDir, tokenFile)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no account connected")
		}
		return fmt.Errorf("deleting token file: %w", err)
	}

	return nil
}

// EnsureValidToken loads the stored token and refreshes it if expired.
// If the refresh fails due to an invalid grant, the token is deleted and
// an error prompting re-login is returned.
func EnsureValidToken(dataDir string, config *oauth2.Config) (*oauth2.Token, error) {
	token, err := LoadToken(dataDir)
	if err != nil {
		return nil, err
	}

	// If the token is still valid, return it directly
	if token.Valid() {
		return token, nil
	}

	// Token is expired, attempt refresh
	if token.RefreshToken == "" {
		// No refresh token available, need to re-authenticate
		if delErr := DeleteToken(dataDir); delErr != nil {
			// Ignore delete errors here
			_ = delErr
		}
		return nil, fmt.Errorf("session expired: run 'gt login' again")
	}

	// Use the oauth2 token source to refresh
	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		// Refresh failed, likely invalid_grant
		if delErr := DeleteToken(dataDir); delErr != nil {
			_ = delErr
		}
		return nil, fmt.Errorf("session expired: run 'gt login' again")
	}

	// Save the refreshed token
	if saveErr := SaveToken(dataDir, newToken); saveErr != nil {
		// Log but don't fail - the token is valid in memory
		fmt.Fprintf(os.Stderr, "warning: could not save refreshed token: %v\n", saveErr)
	}

	return newToken, nil
}

// TokenExists checks whether a token file exists in the data directory.
func TokenExists(dataDir string) bool {
	path := filepath.Join(dataDir, tokenFile)
	_, err := os.Stat(path)
	return err == nil
}

// TokenExpiry returns the expiry time of the stored token, or zero time if
// the token cannot be loaded.
func TokenExpiry(dataDir string) time.Time {
	token, err := LoadToken(dataDir)
	if err != nil {
		return time.Time{}
	}
	return token.Expiry
}
