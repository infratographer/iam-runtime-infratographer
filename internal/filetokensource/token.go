package filetokensource

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// TokenSource implemenets oauth2.TokenSource returning the token from the provided path.
// Loaded tokens are reused
type TokenSource struct {
	mu           sync.Mutex
	path         string
	noReuseToken bool
	token        *oauth2.Token
}

// Token returns the latest token from the configured path.
// Unless Config.NoReuseToken is true, tokens are reused while they're still valid.
func (s *TokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.noReuseToken && s.token != nil && s.token.Valid() {
		return s.token, nil
	}

	tokenb, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("error reading token file: %w", err)
	}

	newToken := string(tokenb)

	// Token signature is not validated here because we only need the expiry time from the claims.
	token, _, err := jwt.NewParser().ParseUnverified(newToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("error parsing jwt: %w", err)
	}

	expiry, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("error getting expiration time from jwt: %w", err)
	}

	var expiryTime time.Time

	if expiry != nil {
		expiryTime = expiry.Time
	}

	s.token = &oauth2.Token{
		AccessToken: newToken,
		TokenType:   "Bearer",
		Expiry:      expiryTime,
	}

	return s.token, nil
}
