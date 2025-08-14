package accesstoken

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/pflag"
	"go.uber.org/multierr"

	"go.infratographer.com/iam-runtime-infratographer/internal/filetokensource"
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
	Source SourceConfig

	// Exchange configures where tokens get exchanges at.
	// If Issuer is empty, token exchange is disabled.
	Exchange ExchangeConfig

	// ExpiryDelta sets early expiry validation for the token.
	// Default is 10 seconds.
	ExpiryDelta time.Duration
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

// SourceConfig configures the source token location for access token exchanges.
// Only one source may be configured at a time.
type SourceConfig struct {
	// File specifies the configuration for sourcing tokens from a file.
	File filetokensource.Config

	// ClientCredentials specifies the oauth2 credentials source the token from.
	ClientCredentials ClientCredentialConfig
}

// Validate ensures the config has been configured properly.
func (c SourceConfig) Validate() error {
	var configured int

	if c.File.Configured() {
		configured++

		if err := c.File.Validate(); err != nil {
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

// ExchangeConfig configures the token exchange provider.
type ExchangeConfig struct {
	// Issuer specifies the URL for the issuer for the exchanged token.
	// The Issuer must support OpenID discovery to discover the token endpoint.
	Issuer string

	// GrantType configures the grant type (default: urn:ietf:params:oauth:grant-type:token-exchange)
	GrantType string

	// TokenType configures the token type (default: urn:ietf:params:oauth:token-type:jwt)
	TokenType string

	// Scopes configures the scopes for the exchanged token.
	Scopes []string
}

func (c ExchangeConfig) configured() bool {
	return c.Issuer != ""
}

// Validate ensures the config has been configured properly.
func (c ExchangeConfig) Validate() error {
	if _, err := url.Parse(c.Issuer); err != nil {
		return err
	}

	return nil
}

// ClientCredentialConfig configures the client credential token source.
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

// AddFlags registers access token flags to the provided flagset.
func AddFlags(flags *pflag.FlagSet) {
	flags.Bool("accessTokenProvider.enabled", false, "enabled configures the access token source for GetAccessToken requests")

	flags.String("accessTokenProvider.source.file.tokenpath", "", "tokenPath is the path to the source jwt token")
	flags.String("accessTokenProvider.source.clientCredentials.issuer", "", "issuer specifies the URL for the issuer for the token request. The Issuer must support OpenID discovery to discover the token endpoint.")
	flags.String("accessTokenProvider.source.clientCredentials.clientID", "", "clientID is the client credentials id which is used to retrieve a token from the issuer. This attribute also supports a file path by prefixing the value with `file://`. example: `file:///var/secrets/client-id`")
	flags.String("accessTokenProvider.source.clientCredentials.clientSecret", "", "clientSecret is the client credentials secret which is used to retrieve a token from the issuer. This attribute also supports a file path by prefixing the value with `file://`. example: `file:///var/secrets/client-secret`")

	flags.String("accessTokenProvider.exchange.issuer", "", "issuer specifies the URL for the issuer for the exchanged token. The Issuer must support OpenID discovery to discover the token endpoint")
	flags.String("accessTokenProvider.exchange.grantType", "urn:ietf:params:oauth:grant-type:token-exchange", "grantType configures the grant type")
	flags.String("accessTokenProvider.exchange.tokenType", "", "tokenType configures the token type")
	flags.StringSlice("accessTokenProvider.exchange.scopes", []string{}, "scopes configures the scopes for the exchanged token")

	flags.Duration("accessTokenProvider.expiryDelta", 10*time.Second, "sets the early expiry validation for the token") //nolint:mnd
}
