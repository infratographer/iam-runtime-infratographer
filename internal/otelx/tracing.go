package otelx

import (
	"go.infratographer.com/x/otelx"
	"go.uber.org/zap"
)

// Initialize sets up OpenTelemetry instrumentation.
func Initialize(config Config, appName string) error {
	otelConfig := otelx.Config{
		Enabled:     config.Enabled,
		Provider:    otelx.ExporterOTLPGRPC,
		Environment: config.Environment,
		SampleRatio: config.SampleRatio,
		OTLP: otelx.OTLPConfig{
			Endpoint: config.URL,
			Insecure: config.Insecure,
		},
	}

	return otelx.InitTracer(otelConfig, appName, zap.NewNop().Sugar())
}
