package eventsx

import "github.com/spf13/pflag"

// Config represents a configuration for events.
type Config struct {
	Enabled bool       `mapstructure:"enabled"`
	NATS    NATSConfig `mapstructure:"nats"`
}

// NATSConfig represents NATS-specific configuration for events.
type NATSConfig struct {
	URL           string
	PublishPrefix string
	PublishTopic  string
	Token         string
	CredsFile     string
}

// AddFlags sets the command line flags for publishing events.
func AddFlags(flags *pflag.FlagSet) {
	flags.Bool("events.enabled", false, "enable NATS event-based functions")
	flags.String("events.nats.url", "", "NATS server URL to use")
	flags.String("events.nats.publishprefix", "", "NATS publish prefix to use")
	flags.String("events.nats.publishtopic", "", "NATS publish topic to use")
	flags.String("events.nats.token", "", "NATS user token to use")
	flags.String("events.nats.credsfile", "", "path to NATS credentials file")
}
