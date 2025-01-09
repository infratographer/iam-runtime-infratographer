package jwt

import (
	"time"

	"github.com/spf13/pflag"
)

// Config represents the configuration for a JWT validator.
type Config struct {
	Disable             bool
	Issuer              string
	JWKSURI             string
	JWKSRefreshInterval time.Duration
}

// AddFlags sets the command line flags for JWT validation.
func AddFlags(flags *pflag.FlagSet) {
	flags.Bool("jwt.disable", false, "Disable JWT service")
	flags.String("jwt.issuer", "", "Issuer to use for JWT validation")
	flags.String("jwt.jwksuri", "", "JWKS URI to use for JWT validation")
	flags.Duration("jwt.jwksrefreshinterval", time.Hour, "sets the jwks refresh interval")
}
