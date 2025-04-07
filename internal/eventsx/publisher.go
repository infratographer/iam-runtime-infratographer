package eventsx

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.infratographer.com/x/events"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const tracerName = "go.infratographer.com/iam-runtime-infratographer/internal/eventsx"

var tracer = otel.GetTracerProvider().Tracer(tracerName)

// Publisher represents something that sends relationships to permissions-api via NATS.
type Publisher interface {
	// PublishAuthRelationship is similar to events.Publisher.PublishAuthRelationship, but with no topic.
	PublishAuthRelationshipRequest(ctx context.Context, message events.AuthRelationshipRequest) (events.Message[events.AuthRelationshipResponse], error)

	// HealthCheck returns nil when the service is healthy.
	HealthCheck(ctx context.Context) error
}

type publisher struct {
	enabled  bool
	topic    string
	innerPub events.AuthRelationshipPublisher
}

func (p publisher) PublishAuthRelationshipRequest(ctx context.Context, message events.AuthRelationshipRequest) (events.Message[events.AuthRelationshipResponse], error) {
	if !p.enabled {
		return nil, ErrPublishNotEnabled
	}

	return p.innerPub.PublishAuthRelationshipRequest(ctx, p.topic, message)
}

// HealthCheck returns nil when the service is healthy.
func (p publisher) HealthCheck(ctx context.Context) error {
	_, span := tracer.Start(ctx, "HealthCheck")
	defer span.End()

	if !p.enabled {
		span.SetAttributes(attribute.String("healthcheck.outcome", "disabled"))

		return nil
	}

	conn := p.innerPub.(*events.NATSConnection).Source().(*nats.Conn)
	if conn.Status() != nats.CONNECTED {
		span.SetStatus(codes.Error, fmt.Sprintf("status not connected: %s", conn.Status()))
		span.SetAttributes(attribute.String("healthcheck.outcome", "unhealthy"))

		return fmt.Errorf("%w: status: %s", ErrPublisherNotConnected, conn.Status())
	}

	span.SetAttributes(attribute.String("healthcheck.outcome", "healthy"))

	return nil
}

// NewPublisher creates a new events publisher from the given config.
func NewPublisher(cfg Config) (Publisher, error) {
	if !cfg.Enabled {
		return publisher{
			enabled:  false,
			topic:    "",
			innerPub: nil,
		}, nil
	}

	natsCfg := events.NATSConfig{
		URL:           cfg.NATS.URL,
		PublishPrefix: cfg.NATS.PublishPrefix,
		Token:         cfg.NATS.Token,
		CredsFile:     cfg.NATS.CredsFile,
	}

	conn, err := events.NewNATSConnection(natsCfg)
	if err != nil {
		return nil, err
	}

	out := publisher{
		enabled:  true,
		topic:    cfg.NATS.PublishTopic,
		innerPub: conn,
	}

	return out, nil
}
