package accesstoken

import (
	"errors"
	"fmt"
	"net/url"

	"go.infratographer.com/iam-runtime-infratographer/internal/filetokensource"
	"go.uber.org/multierr"
)

var (
	// ErrTooManyAccessTokenSources is returned when more than one access token source has been configured.
	ErrTooManyAccessTokenSources = errors.New("only one access token source may be configured")

	// ErrNoAccessTokenSources is returned when no access token sources have been configured.
	ErrNoAccessTokenSources = errors.New("no access token sources have been configured")

	// ErrClientCredentialIssuerRequired is returned when no Issuer has been configured.
	ErrClientCredentialIssuerRequired = errors.New("issuer is required")

	// ErrClientCredentialClientIDRequired is returned when no ClientID has been configured.
	ErrClientCredentialClientIDRequired = errors.New("clientID is required")

	// ErrClientCredentialClientSecretRequired is returned when no ClientSecret has been configured.
	ErrClientCredentialClientSecretRequired = errors.New("clientSecret is required")
)

// Config defines the configuration for sourcing a token.
// Source defines the location to retrieve a token from.
// If Exchange has been configured, the source token will be exchanged
// with the defined exchange and will return a new token upon request.
// If Exchange has not been configured, the token from the source token
// provider is simply returned.
type Config struct {
	// Enabled configures the access token source for GetAccessToken requests.
	Enabled bool

	// Source configures the location to source tokens from.
	Source AccessTokenSourceConfig

	// Exchange configures where tokens get exchanges at.
	// If Issuer is empty, token exchange is disabled.
	Exchange AccessTokenExchangeConfig
}

// Validate ensures the config has been configured properly.
func (c Config) Validate() error {
	var errs error

	if err := c.Source.Validate(); err != nil {
		errs = multierr.Append(errs, fmt.Errorf("source: %w", err))
	}

	if c.Exchange.configured() {
		if err := c.Exchange.Validate(); err != nil {
			errs = multierr.Append(errs, fmt.Errorf("exchange: %w", err))
		}
	}

	return errs
}

// AccessTokenSourceConfig configures the source token location for access token exchanges.
// Only one source may be configured at a time.
type AccessTokenSourceConfig struct {
	// FileToken specifies the configuration for sourcing tokens from a file.
	FileToken filetokensource.Config

	// ClientCredentials specifies the oauth2 credentials source the token from.
	ClientCredentials ClientCredentialConfig
}

// Validate ensures the config has been configured properly.
func (c AccessTokenSourceConfig) Validate() error {
	var configured int

	if c.FileToken.Configured() {
		configured++

		if err := c.FileToken.Validate(); err != nil {
			return fmt.Errorf("fileToken: %w", err)
		}
	}

	if c.ClientCredentials.configured() {
		configured++

		if err := c.ClientCredentials.Validate(); err != nil {
			return fmt.Errorf("clientCredentials: %w", err)
		}
	}

	switch configured {
	case 0:
		return ErrNoAccessTokenSources
	case 1:
	default:
		return ErrTooManyAccessTokenSources
	}

	return nil
}

// AccessTokenExchangeConfig configures the token exchange provider.
type AccessTokenExchangeConfig struct {
	// Issuer specifies the URL for the issuer for the exchanged token.
	// The Issuer must support OpenID discovery to discover the token endpoint.
	Issuer string

	// GrantType configures the grant type (default: urn:ietf:params:oauth:grant-type:token-exchange)
	GrantType string

	// TokenType configures the token type (default: urn:ietf:params:oauth:token-type:jwt)
	TokenType string
}

func (c AccessTokenExchangeConfig) configured() bool {
	return c.Issuer != ""
}

// Validate ensures the config has been configured properly.
func (c AccessTokenExchangeConfig) Validate() error {
	if _, err := url.Parse(c.Issuer); err != nil {
		return err
	}

	return nil
}

// ClientCredential configures the client credential token source.
type ClientCredentialConfig struct {
	// Issuer specifies the URL for the issuer for the token request.
	// The Issuer must support OpenID discovery to discover the token endpoint.
	Issuer string

	// ClientID is the client credentials id which is used to retrieve a token from the issuer.
	ClientID string

	// ClientSecret is the client credentials secret which is used to retrieve a token from the issuer.
	ClientSecret string
}

func (c ClientCredentialConfig) configured() bool {
	return c.Issuer != "" || c.ClientID != "" || c.ClientSecret != ""
}

// Validate ensures the config has been configured properly.
func (c ClientCredentialConfig) Validate() error {
	if c.Issuer == "" {
		return ErrClientCredentialIssuerRequired
	}

	if c.ClientID == "" {
		return ErrClientCredentialClientIDRequired
	}

	if c.ClientSecret == "" {
		return ErrClientCredentialClientSecretRequired
	}

	return nil
}
