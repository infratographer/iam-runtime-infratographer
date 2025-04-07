package permissions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go.infratographer.com/iam-runtime-infratographer/internal/selecthost"
)

const (
	apiRoute         = "/api/v1/allow"
	healthCheckRoute = "/readyz"

	contentTypeApplicationJSON = "application/json"

	headerAuthorization = "Authorization"
	headerContentType   = "Content-Type"

	prefixBearer = "Bearer "

	outcomeAllowed = "allowed"
	outcomeDenied  = "denied"

	tracerName = "go.infratographer.com/iam-runtime-infratographer/internal/permissions"

	clientTimeout = 5 * time.Second
)

// RequestAction represents an (action, resource) pair to check access to in a request.
type RequestAction struct {
	Action     string `json:"action"`
	ResourceID string `json:"resource_id"`
}

type checkPermissionRequest struct {
	Actions []RequestAction `json:"actions"`
}

// Client represents a client for interacting with permissions-api.
type Client interface {
	CheckAccess(ctx context.Context, subjToken string, actions []RequestAction) error

	// HealthCheck returns nil when the service is healthy.
	HealthCheck(ctx context.Context) error
}

type client struct {
	enabled        bool
	apiURL         string
	healthCheckURL string
	httpClient     *retryablehttp.Client
	tracer         trace.Tracer
	logger         *zap.SugaredLogger
}

// NewClient creates a new permissions-api client.
func NewClient(config Config, logger *zap.SugaredLogger) (Client, error) {
	if config.Disable {
		return &client{
			enabled: false,
			tracer:  otel.GetTracerProvider().Tracer(tracerName),
			logger:  logger,
		}, nil
	}

	apiURLString := fmt.Sprintf("https://%s%s", config.Host, apiRoute)

	if _, err := url.Parse(apiURLString); err != nil {
		return nil, err
	}

	transport, err := config.initTransport(http.DefaultTransport, selecthost.Logger(logger))
	if err != nil {
		return nil, err
	}

	httpClient := retryablehttp.NewClient()

	httpClient.RetryWaitMin = 100 * time.Millisecond //nolint:mnd
	httpClient.RetryWaitMax = 2 * time.Second        //nolint:mnd
	httpClient.Logger = &retryableLogger{logger}
	httpClient.HTTPClient = &http.Client{
		Timeout:   clientTimeout,
		Transport: transport,
	}

	out := &client{
		enabled:        true,
		apiURL:         apiURLString,
		healthCheckURL: fmt.Sprintf("https://%s%s", config.Host, healthCheckRoute),
		httpClient:     httpClient,
		tracer:         otel.GetTracerProvider().Tracer(tracerName),
		logger:         logger,
	}

	return out, nil
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= http.StatusMultiStatus {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return ErrUnauthenticated
		case http.StatusForbidden:
			return ErrPermissionDenied
		default:
			return fmt.Errorf("%w: status code %d", ErrUnexpectedResponse, resp.StatusCode)
		}
	}

	return nil
}

func (c *client) CheckAccess(ctx context.Context, subjToken string, actions []RequestAction) error {
	ctx, span := c.tracer.Start(ctx, "CheckAccess")
	defer span.End()

	if !c.enabled {
		span.SetStatus(codes.Error, ErrServiceDisabled.Error())
		c.logger.Error("CheckAccess called but service is disabled")

		return ErrServiceDisabled
	}

	request := checkPermissionRequest{
		Actions: actions,
	}

	var reqBody bytes.Buffer

	// Marshal the request body based on the provided actions.
	if err := json.NewEncoder(&reqBody).Encode(request); err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.logger.Errorw("failed to encode permissions-api request body", "error", err)

		return err
	}

	// Build the request to send up to permissions-api.
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, &reqBody)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.logger.Errorw("failed to create permissions-api request", "error", err)

		return err
	}

	// Pass the token provided to the client directly up as a bearer token.
	authHeader := prefixBearer + subjToken

	req.Header.Set(headerAuthorization, authHeader)
	req.Header.Set(headerContentType, contentTypeApplicationJSON)

	// If for some reason we fail to send the request, bail.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.logger.Errorw("failed to make permissions-api request", "error", err)

		return err
	}

	defer resp.Body.Close() //nolint:errcheck

	// Check what the outcome of the request was; if it was fine, terminate early.
	err = checkResponse(resp)
	if err == nil {
		span.SetAttributes(
			attribute.String(
				"permissions.outcome",
				outcomeAllowed,
			),
		)

		return nil
	}

	body, readErr := io.ReadAll(resp.Body) //nolint:errcheck
	if readErr != nil {
		c.logger.Warnw("error reading permissions-api response body", "error", readErr.Error())
	}

	// A 401 is a failure state as far as the client is concerned; a 403 is not. Update the span we
	// send up accordingly.
	switch {
	case errors.Is(err, ErrUnauthenticated):
		span.SetStatus(codes.Error, err.Error())
	case errors.Is(err, ErrPermissionDenied):
		span.AddEvent("permission denied")
		span.SetAttributes(
			attribute.String(
				"permissions.outcome",
				outcomeDenied,
			),
		)
	case errors.Is(err, ErrUnexpectedResponse):
		c.logger.Errorw("unexpected response from server", "error", err, "response.status_code", resp.StatusCode, "response.body", string(body))
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

// HealthCheck returns nil when the service is healthy.
func (c *client) HealthCheck(ctx context.Context) error {
	ctx, span := c.tracer.Start(ctx, "HealthCheck")
	defer span.End()

	if !c.enabled {
		span.SetAttributes(attribute.String("healthcheck.outcome", "disabled"))

		return nil
	}

	span.SetAttributes(attribute.String("healthcheck.outcome", "unhealthy"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.healthCheckURL, nil)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	resp, err := c.httpClient.HTTPClient.Do(req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	defer resp.Body.Close() //nolint:errcheck

	_, _ = io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%w: status code: %d", ErrUnexpectedResponse, resp.StatusCode)

		span.SetStatus(codes.Error, err.Error())

		return err
	}

	span.SetAttributes(attribute.String("healthcheck.outcome", "healthy"))

	return nil
}
