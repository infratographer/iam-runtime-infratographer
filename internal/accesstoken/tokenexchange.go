package accesstoken

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.infratographer.com/iam-runtime-infratographer/internal/jwt"
	"golang.org/x/oauth2"
)

var (
	// TokenExchangeError is a root error for all other token exchange errors.
	TokenExchangeError = errors.New("failed to exchange token") //nolint:revive,stylecheck // not returned directly, but used as a root error.

	// ErrUpstreamTokenRequestFailed is returned when the upstream token provider returns an error.
	ErrUpstreamTokenRequestFailed = fmt.Errorf("%w, upstream token request failed", TokenExchangeError)

	// ErrInvalidTokenExchangeRequest is returned when the request returns a status 400 BadRequest.
	ErrInvalidTokenExchangeRequest = fmt.Errorf("%w, invalid request", TokenExchangeError)

	// ErrTokenExchangeRequestFailed is returned when an error is generated while exchanging the token.
	ErrTokenExchangeRequestFailed = fmt.Errorf("%w, failed request", TokenExchangeError)
)

const (
	defaultGrantType = "urn:ietf:params:oauth:grant-type:token-exchange"
	defaultTokenType = "urn:ietf:params:oauth:token-type:jwt"
)

type exchangeTokenSource struct {
	cfg            AccessTokenExchangeConfig
	ctx            context.Context
	upstream       oauth2.TokenSource
	exchangeConfig oauth2.Config
}

// Token retrieves an OAuth 2.0 access token from the configured issuer using token exchange.
func (s *exchangeTokenSource) Token() (*oauth2.Token, error) {
	upstreamToken, err := s.upstream.Token()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamTokenRequestFailed, err)
	}

	token, err := s.exchangeConfig.Exchange(s.ctx, "",
		oauth2.SetAuthURLParam("grant_type", s.cfg.GrantType),
		oauth2.SetAuthURLParam("subject_token", upstreamToken.AccessToken),
		oauth2.SetAuthURLParam("subject_token_type", s.cfg.TokenType),
	)
	if err != nil {
		if rErr, ok := err.(*oauth2.RetrieveError); ok {
			if rErr.Response.StatusCode == http.StatusBadRequest {
				return nil, fmt.Errorf("%w: %w", ErrInvalidTokenExchangeRequest, rErr)
			}
		}

		return nil, fmt.Errorf("%w: %w", ErrTokenExchangeRequestFailed, err)
	}

	return token, nil
}

func newExchangeTokenSource(ctx context.Context, cfg AccessTokenExchangeConfig, upstream oauth2.TokenSource) (oauth2.TokenSource, error) {
	tokenEndpoint, err := jwt.FetchIssuerTokenEndpoint(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange issuer token endpoint: %w", err)
	}

	if cfg.GrantType == "" {
		cfg.GrantType = defaultGrantType
	}

	if cfg.TokenType == "" {
		cfg.TokenType = defaultTokenType
	}

	return &exchangeTokenSource{
		cfg:      cfg,
		ctx:      ctx,
		upstream: upstream,
		exchangeConfig: oauth2.Config{
			Endpoint: oauth2.Endpoint{
				TokenURL: tokenEndpoint,
			},
		},
	}, nil
}
