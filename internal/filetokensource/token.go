package filetokensource

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// TokenSource implemenets oauth2.TokenSource returning the token from the provided path.
type TokenSource struct {
	path string
}

// Token returns the latest token from the configured path.
func (s *TokenSource) Token() (*oauth2.Token, error) {
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

	return &oauth2.Token{
		AccessToken: newToken,
		TokenType:   "Bearer",
		Expiry:      expiryTime,
	}, nil
}
