package accesstoken

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"go.infratographer.com/iam-runtime-infratographer/internal/jwt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func (c Config) toTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	source, err := c.Source.toTokenSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("token source: %w", err)
	}

	if c.Exchange.configured() {
		source, err = c.Exchange.toTokenSource(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("token exchange: %w", err)
		}
	}

	return source, nil
}

func (c AccessTokenSourceConfig) toTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	if c.FileToken.Configured() {
		tokensource, err := c.FileToken.ToTokenSource()
		if err != nil {
			return nil, fmt.Errorf("file token: %w", err)
		}

		return tokensource, nil
	}

	tokensource, err := c.ClientCredentials.toTokenSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("client credentials: %w", err)
	}

	return tokensource, nil
}

func (c ClientCredentialConfig) toTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	tokenEndpoint, err := jwt.FetchIssuerTokenEndpoint(ctx, c.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issuer token endpoint: %w", err)
	}

	clientID := c.ClientID
	clientSecret := c.ClientSecret

	if uri, err := url.ParseRequestURI(clientID); err == nil && uri.Scheme == "file" {
		file := filepath.Join(uri.Host, uri.Path)

		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		clientID = strings.TrimSpace(string(content))
	}

	if uri, err := url.ParseRequestURI(clientSecret); err == nil && uri.Scheme == "file" {
		file := filepath.Join(uri.Host, uri.Path)

		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		clientSecret = strings.TrimSpace(string(content))
	}

	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenEndpoint,
	}

	return config.TokenSource(ctx), nil
}

func (c AccessTokenExchangeConfig) toTokenSource(ctx context.Context, upstream oauth2.TokenSource) (oauth2.TokenSource, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return newExchangeTokenSource(ctx, c, upstream)
}

// NewTokenSource initializes a new token source from the provided config.
// If the config has Enabled false, then a disabled token source is returned.
func NewTokenSource(ctx context.Context, cfg Config) (oauth2.TokenSource, error) {
	if !cfg.Enabled {
		return &disabledTokenSource{}, nil
	}

	return cfg.toTokenSource(ctx)
}
