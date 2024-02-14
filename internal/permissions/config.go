package permissions

import (
	"github.com/spf13/pflag"
)

// Config represents a permissions-api client configuration.
type Config struct {
	// Host represents a permissions-api host to hit.
	Host string
}

// AddFlags sets the command line flags for the permissions-api client.
func AddFlags(flags *pflag.FlagSet) {
	flags.String("permissions.host", "", "permissions-api host to use")
}
