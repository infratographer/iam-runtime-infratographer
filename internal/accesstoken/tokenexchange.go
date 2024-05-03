package accesstoken

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

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
	mu             sync.Mutex
	upstream       oauth2.TokenSource
	upstreamToken  *oauth2.Token
	exchangeConfig oauth2.Config
	token          *oauth2.Token
	err            error
}

// Token retrieves an OAuth 2.0 access token from the configured issuer using token exchange.
// Tokens are reused as long as they are valid.
// Upstream tokens used as the source for the exchange are reused as long as they are valid.
func (s *exchangeTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != nil && s.token.Valid() {
		return s.token, nil
	}

	if err := s.refreshUpstream(); err != nil {
		s.err = err

		return s.token, err
	}

	if err := s.exchange(); err != nil {
		s.err = err

		return s.token, err
	}

	s.err = nil

	return s.token, nil
}

func (s *exchangeTokenSource) refreshUpstream() error {
	if s.upstreamToken == nil || !s.upstreamToken.Valid() {
		token, err := s.upstream.Token()
		if err != nil {
			s.err = err

			return fmt.Errorf("%w: %w", ErrUpstreamTokenRequestFailed, err)
		}

		s.upstreamToken = token
	}

	return nil
}

func (s *exchangeTokenSource) exchange() error {
	token, err := s.exchangeConfig.Exchange(s.ctx, "",
		oauth2.SetAuthURLParam("grant_type", s.cfg.GrantType),
		oauth2.SetAuthURLParam("subject_token", s.upstreamToken.AccessToken),
		oauth2.SetAuthURLParam("subject_token_type", s.cfg.TokenType),
	)
	if err != nil {
		if rErr, ok := err.(*oauth2.RetrieveError); ok {
			if rErr.Response.StatusCode == http.StatusBadRequest {
				return fmt.Errorf("%w: %w", ErrInvalidTokenExchangeRequest, rErr)
			}
		}

		return fmt.Errorf("%w: %w", ErrTokenExchangeRequestFailed, err)
	}

	s.token = token

	return nil
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
