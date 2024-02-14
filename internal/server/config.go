package server

import (
	"github.com/spf13/pflag"
)

// Config represents a configuration for an IAM runtime server.
type Config struct {
	SocketPath string
}

// AddFlags sets the command line flags for the IAM runtime server.
func AddFlags(flags *pflag.FlagSet) {
	flags.String("server.socketpath", "", "gRPC server socket path")
}
