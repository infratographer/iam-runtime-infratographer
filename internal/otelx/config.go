package otelx

import (
	"github.com/spf13/pflag"
)

// Config represents an OpenTelemetry configuration.
type Config struct {
	Enabled  bool
	URL      string
	Insecure bool
}

// AddFlags sets the command line flags for OpenTelemetry instrumentation.
func AddFlags(flags *pflag.FlagSet) {
	flags.Bool("tracing.enabled", false, "true if tracing should be enabled")
	flags.String("tracing.url", "", "gRPC URL for OpenTelemetry collector")
	flags.Bool("tracing.insecure", false, "true if TLS should be disabled")
}
