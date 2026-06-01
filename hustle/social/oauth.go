package social

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"time"
)

// OAuthState manages the authentication lifecycle for social providers
type OAuthState struct {
	Provider     string
	ClientID     string
	ClientSecret string
	Token        *oauth2.Token
	Config       *oauth2.Config
}

// NewOAuthState creates a scaffold for OAuth2 flow
func NewOAuthState(provider, clientID, clientSecret string, scopes []string, authURL, tokenURL string) *OAuthState {
	return &OAuthState{
		Provider:     provider,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
		},
	}
}

// GetToken returns a valid token, refreshing if necessary
func (s *OAuthState) GetToken(ctx context.Context) (*oauth2.Token, error) {
	if s.Token == nil {
		return nil, fmt.Errorf("no token available; manual authorization required")
	}

	ts := s.Config.TokenSource(ctx, s.Token)
	tok, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}

	// Update local token if it was refreshed
	if tok.AccessToken != s.Token.AccessToken {
		fmt.Printf("[OAuth] %s token refreshed. Expiry: %s\n", s.Provider, tok.Expiry.Format(time.RFC3339))
		s.Token = tok
	}

	return tok, nil
}

// Exchange converts an authorization code into a token
func (s *OAuthState) Exchange(ctx context.Context, code string) error {
	tok, err := s.Config.Exchange(ctx, code)
	if err != nil {
		return err
	}
	s.Token = tok
	fmt.Printf("[OAuth] Successfully exchanged code for %s token\n", s.Provider)
	return nil
}
