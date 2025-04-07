package server

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

const serviceHealthCheckTimeout = 30 * time.Second

// HealthChecker defines a health checker service.
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// HealthChecks specifies a service name and the health check function to call.
type HealthChecks map[string]HealthChecker

var _ health.HealthServer = (*server)(nil)

// Check implements Health service checks.
func (s *server) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	span := trace.SpanFromContext(ctx)

	s.logger.Info("received HealthCheck request")

	status := health.HealthCheckResponse_SERVING

	if s.healthChecks != nil {
		for name, svc := range s.healthChecks {
			if in.GetService() != "" && in.GetService() != name {
				continue
			}

			chkCtx, cancel := context.WithTimeout(ctx, serviceHealthCheckTimeout)
			defer cancel()

			if err := svc.HealthCheck(chkCtx); err == nil {
				s.logger.Debugf("Service %s healthy", name)
			} else {
				s.logger.Errorf("Service %s unhealthy: %s", name, err.Error())

				status = health.HealthCheckResponse_NOT_SERVING

				break
			}
		}
	}

	span.SetAttributes(
		attribute.String("healthcheck.outcome", status.String()),
	)

	resp := &health.HealthCheckResponse{
		Status: status,
	}

	return resp, nil
}
