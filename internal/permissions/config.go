package permissions

import (
	"net/http"
	"time"

	"github.com/spf13/pflag"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.infratographer.com/iam-runtime-infratographer/internal/selecthost"
)

// Config represents a permissions-api client configuration.
type Config struct {
	// Disable disables the permissions service.
	Disable bool

	// Host represents a permissions-api host to hit.
	Host string

	// Discovery defines the host discovery configuration.
	Discovery DiscoveryConfig
}

func (c Config) initTransport(base http.RoundTripper, opts ...selecthost.Option) (http.RoundTripper, error) {
	base = otelhttp.NewTransport(base)

	if c.Disable || c.Discovery.Disable {
		return base, nil
	}

	cOpts := []selecthost.Option{
		selecthost.Fallback(c.Host),
	}

	discovery := c.Discovery

	if discovery.Interval > 0 {
		cOpts = append(cOpts, selecthost.DiscoveryInterval(discovery.Interval))
	}

	if discovery.Quick != nil && *discovery.Quick {
		cOpts = append(cOpts, selecthost.Quick())
	}

	if discovery.Optional == nil || *discovery.Optional {
		cOpts = append(cOpts, selecthost.Optional())
	}

	if discovery.Prefer != "" {
		cOpts = append(cOpts, selecthost.Prefer(discovery.Prefer))
	}

	if discovery.Fallback != "" {
		cOpts = append(cOpts, selecthost.Fallback(discovery.Fallback))
	}

	check := discovery.Check

	if check.Scheme != "" {
		cOpts = append(cOpts, selecthost.CheckScheme(check.Scheme))
	}

	if check.Path != "" {
		cOpts = append(cOpts, selecthost.CheckPath(check.Path))
	} else {
		cOpts = append(cOpts, selecthost.CheckPath("/readyz"))
	}

	if check.Count > 0 {
		cOpts = append(cOpts, selecthost.CheckCount(check.Count))
	}

	if check.Interval > 0 {
		cOpts = append(cOpts, selecthost.CheckInterval(check.Interval))
	}

	if check.Delay > 0 {
		cOpts = append(cOpts, selecthost.CheckDelay(check.Delay))
	}

	if check.Timeout > 0 {
		cOpts = append(cOpts, selecthost.CheckTimeout(check.Timeout))
	}

	if check.Concurrency > 0 {
		cOpts = append(cOpts, selecthost.CheckConcurrency(check.Concurrency))
	}

	selector, err := selecthost.NewSelector(c.Host, "permissions-api", "tcp", append(cOpts, opts...)...)
	if err != nil {
		return nil, err
	}

	selector.Start()

	return selecthost.NewTransport(selector, base), nil
}

// DiscoveryConfig represents the host discovery configuration.
type DiscoveryConfig struct {
	// Disable disables host discovery.
	//
	// Default: false
	Disable bool

	// Interval sets the frequency at which SRV records are rediscovered.
	//
	// Default: 15m
	Interval time.Duration

	// Quick ensures a quick startup, allowing for a more optimal host to be chosen after discovery has occurred.
	// When Quick is enabled, the default fallback address or default host is immediately returned.
	// Once the discovery process has completed, a discovered host will be selected.
	//
	// Default: false
	Quick *bool

	// Optional uses the fallback address or default host without throwing errors.
	// The discovery process continues to run in the background, in the chance that SRV records are added at a later point.
	//
	// Default: true
	Optional *bool

	// Check customizes the target health checking process.
	Check CheckConfig

	// Prefer specifies a preferred host.
	// If the host is not discovered or has an error, it will not be used.
	Prefer string

	// Fallback specifies a fallback host if no hosts are discovered or all hosts are currently failing.
	//
	// Default: [Config] Host
	Fallback string
}

// CheckConfig defines the configuration for host checks.
type CheckConfig struct {
	// Scheme sets the check URI scheme.
	// Default is http unless discovered host port is 443 in which scheme is th en https
	Scheme string

	// Path sets the request path for checks.
	//
	// Default: /readyz
	Path string

	// Count defines the number of checks to run on each endpoint.
	//
	// Default: 5
	Count int

	// Interval specifies how frequently to run checks.
	//
	// Default: 1m
	Interval time.Duration

	// Delay specifies how long to wait between subsequent checks for the same host.
	//
	// Default: 200ms
	Delay time.Duration

	// Timeout defines the maximum time an individual check request can take.
	//
	// Default: 2s
	Timeout time.Duration

	// Concurrency defines the number of hosts which may be checked simultaneously.
	//
	// Default: 5
	Concurrency int
}

// AddFlags sets the command line flags for the permissions-api client.
func AddFlags(flags *pflag.FlagSet) {
	flags.Bool("permissions.disable", false, "disables permissions service")
	flags.String("permissions.host", "", "permissions-api host to use")
}
