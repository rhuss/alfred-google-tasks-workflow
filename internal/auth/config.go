package auth

import "golang.org/x/oauth2"

// OAuthConfig bundles a validated OAuth config and token for convenience.
type OAuthConfig struct {
	Config *oauth2.Config
	Token  *oauth2.Token
}
