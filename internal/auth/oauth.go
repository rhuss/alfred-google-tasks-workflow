package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const (
	oauthTimeout = 3 * time.Minute
)

// successHTML is shown in the browser after successful OAuth authentication.
const successHTML = `<!DOCTYPE html>
<html>
<head><title>Google Tasks - Authenticated</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; display: flex;
  justify-content: center; align-items: center; height: 100vh; margin: 0;
  background: #f5f5f5; color: #333; }
.card { background: white; padding: 2em 3em; border-radius: 12px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center; }
h1 { color: #4285f4; margin-bottom: 0.5em; }
</style></head>
<body>
<div class="card">
  <h1>Authenticated!</h1>
  <p>You can close this tab and return to Alfred.</p>
</div>
</body>
</html>`

// errorHTML is shown in the browser when OAuth fails.
const errorHTML = `<!DOCTYPE html>
<html>
<head><title>Google Tasks - Error</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; display: flex;
  justify-content: center; align-items: center; height: 100vh; margin: 0;
  background: #f5f5f5; color: #333; }
.card { background: white; padding: 2em 3em; border-radius: 12px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center; }
h1 { color: #d93025; margin-bottom: 0.5em; }
</style></head>
<body>
<div class="card">
  <h1>Authentication Failed</h1>
  <p>%s</p>
  <p>Please try again from Alfred with <code>gt login</code>.</p>
</div>
</body>
</html>`

// generateCodeVerifier creates a random PKCE code verifier (43-128 chars).
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating code verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge creates the S256 code challenge from the verifier.
func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// RunOAuthFlow starts the OAuth 2.0 authorization code flow with PKCE.
// It opens the browser to Google's consent screen, starts a local HTTP server
// to capture the callback, and exchanges the authorization code for tokens.
func RunOAuthFlow(config *oauth2.Config) (*oauth2.Token, error) {
	// Generate PKCE parameters
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return nil, err
	}
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Start a local HTTP server on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("starting local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Update config with the actual redirect URL
	config.RedirectURL = redirectURL

	// Generate a random state parameter for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		listener.Close()
		return nil, fmt.Errorf("generating state parameter: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	// Build the authorization URL with PKCE
	authURL := config.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	// Channel to receive the auth code or error
	type authResult struct {
		code string
		err  error
	}
	resultCh := make(chan authResult, 1)

	// Set up the callback handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		queryErr := r.URL.Query().Get("error")
		if queryErr != "" {
			errDesc := r.URL.Query().Get("error_description")
			if errDesc == "" {
				errDesc = queryErr
			}
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, errorHTML, errDesc)
			resultCh <- authResult{err: fmt.Errorf("OAuth error: %s", errDesc)}
			return
		}

		// Verify state parameter to prevent CSRF attacks
		returnedState := r.URL.Query().Get("state")
		if returnedState != state {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, errorHTML, "Invalid state parameter (possible CSRF attack)")
			resultCh <- authResult{err: fmt.Errorf("state parameter mismatch")}
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, errorHTML, "No authorization code received")
			resultCh <- authResult{err: fmt.Errorf("no authorization code received")}
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, successHTML)
		resultCh <- authResult{code: code}
	})

	server := &http.Server{Handler: mux}

	// Start server in background
	go func() {
		if serveErr := server.Serve(listener); serveErr != http.ErrServerClosed {
			resultCh <- authResult{err: fmt.Errorf("local server error: %w", serveErr)}
		}
	}()

	// Open browser (will be called by the workflow command handler)
	// Return the URL for the caller to open
	fmt.Printf("Opening browser for authentication...\n")
	if browserErr := openBrowser(authURL); browserErr != nil {
		listener.Close()
		return nil, fmt.Errorf("opening browser: %w", browserErr)
	}

	// Wait for the callback or timeout
	var result authResult
	select {
	case result = <-resultCh:
	case <-time.After(oauthTimeout):
		result = authResult{err: fmt.Errorf("authentication timed out after %v - please try again", oauthTimeout)}
	}

	// Shut down the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	if result.err != nil {
		return nil, result.err
	}

	// Exchange the authorization code for tokens (with PKCE verifier)
	token, err := config.Exchange(context.Background(), result.code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("exchanging authorization code: %w", err)
	}

	return token, nil
}

// openBrowser opens the given URL in the default browser on macOS.
func openBrowser(url string) error {
	return execCommand("open", url)
}
