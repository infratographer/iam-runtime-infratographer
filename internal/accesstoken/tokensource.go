package accesstoken

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"go.infratographer.com/iam-runtime-infratographer/internal/jwt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const tracerName = "go.infratographer.com/iam-runtime-infratographer/internal/accesstoken"

var tracer = otel.GetTracerProvider().Tracer(tracerName)

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

	source = oauth2.ReuseTokenSourceWithExpiry(nil, source, c.ExpiryDelta)

	return source, nil
}

func (c AccessTokenSourceConfig) toTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	if c.File.Configured() {
		tokensource, err := c.File.ToTokenSource()
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

// HealthyTokenSource extends oauth2.TokenSource implementing the HealthChecker interface.
type HealthyTokenSource interface {
	oauth2.TokenSource

	// HealthCheck returns nil when the service is healthy.
	HealthCheck(ctx context.Context) error
}

type healthyTokenSource struct {
	oauth2.TokenSource
}

// HealthCheck returns nil when the service is healthy.
func (s *healthyTokenSource) HealthCheck(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "HealthCheck")
	defer span.End()

	errCh := make(chan error, 1)

	go func() {
		_, err := s.TokenSource.Token()
		errCh <- err

		close(errCh)
	}()

	select {
	case <-ctx.Done():
		span.SetStatus(codes.Error, ctx.Err().Error())
		span.RecordError(ctx.Err())
		span.SetAttributes(attribute.String("healthcheck.outcome", "unhealthy"))

		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			if errors.Is(err, ErrAccessTokenSourceNotEnabled) {
				span.SetAttributes(attribute.String("healthcheck.outcome", "disabled"))

				return nil
			} else {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				span.SetAttributes(attribute.String("healthcheck.outcome", "unhealthy"))

				return err
			}
		}
	}

	span.SetAttributes(attribute.String("healthcheck.outcome", "healthy"))

	return nil
}

// NewTokenSource initializes a new token source from the provided config.
// If the config has Enabled false, then a disabled token source is returned.
func NewTokenSource(ctx context.Context, cfg Config) (HealthyTokenSource, error) {
	if !cfg.Enabled {
		return &healthyTokenSource{&disabledTokenSource{}}, nil
	}

	ts, err := cfg.toTokenSource(ctx)
	if err != nil {
		return nil, err
	}

	return &healthyTokenSource{ts}, nil
}
