package otelx

import (
	"context"
	"net/url"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	timeout = 10 * time.Second
)

func initializeExporter(config Config) (trace.SpanExporter, error) {
	_, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.URL),
		otlptracegrpc.WithTimeout(timeout),
	}

	if config.Insecure {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
	}

	return otlptrace.New(context.Background(), otlptracegrpc.NewClient(exporterOpts...))
}

// Initialize sets up OpenTelemetry instrumentation.
func Initialize(config Config, appName string) error {
	providerOpts := []trace.TracerProviderOption{
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(appName),
		)),
	}

	if config.Enabled {
		exporter, err := initializeExporter(config)
		if err != nil {
			return err
		}

		providerOpts = append(providerOpts, trace.WithBatcher(exporter))
	}

	tp := trace.NewTracerProvider(providerOpts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return nil
}
