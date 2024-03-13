package config

import (
	"go.infratographer.com/iam-runtime-infratographer/internal/eventsx"
	"go.infratographer.com/iam-runtime-infratographer/internal/jwt"
	"go.infratographer.com/iam-runtime-infratographer/internal/otelx"
	"go.infratographer.com/iam-runtime-infratographer/internal/permissions"
	"go.infratographer.com/iam-runtime-infratographer/internal/server"
)

// Config represents a configuration for iam-runtime-infratographer.
type Config struct {
	JWT         jwt.Config
	Permissions permissions.Config
	Events      eventsx.Config
	Server      server.Config
	Tracing     otelx.Config
}
