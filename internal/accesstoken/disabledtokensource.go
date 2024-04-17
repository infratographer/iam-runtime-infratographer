package accesstoken

import (
	"errors"

	"golang.org/x/oauth2"
)

// ErrAccessTokenSourceNotEnabled is returned when a token is requested for the token source which has not been enabled.
var ErrAccessTokenSourceNotEnabled = errors.New("access token source is not enabled")

type disabledTokenSource struct{}

// Token always returns an error that the token source is not enabled.
func (s disabledTokenSource) Token() (*oauth2.Token, error) {
	return nil, ErrAccessTokenSourceNotEnabled
}
