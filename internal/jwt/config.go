package jwt

import (
	"github.com/spf13/pflag"
)

// Config represents the configuration for a JWT validator.
type Config struct {
	Issuer  string
	JWKSURI string
}

// AddFlags sets the command line flags for JWT validation.
func AddFlags(flags *pflag.FlagSet) {
	flags.String("jwt.issuer", "", "Issuer to use for JWT validation")
	flags.String("jwt.jwksuri", "", "JWKS URI to use for JWT validation")
}
