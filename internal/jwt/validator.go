package jwt

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const tracerName = "go.infratographer.com/iam-runtime-infratographer/internal/jwt"

var (
	// ErrIssuerKeysMissing is returned by the health check when no issuer keys exist in the store.
	ErrIssuerKeysMissing = errors.New("issuer keys missing")

	tracer = otel.GetTracerProvider().Tracer(tracerName)
)

// Validator represents a JWT validator.
type Validator interface {
	// ValidateToken checks that the given token is valid (i.e., is well-formed with a valid
	// signature and future expiry). On success, it returns a map of claims describing the subject.
	ValidateToken(string) (string, map[string]any, error)

	// HealthCheck returns nil when the service is healthy.
	HealthCheck(ctx context.Context) error
}

type validator struct {
	kf     jwt.Keyfunc
	parser *jwt.Parser

	keyStorage jwkset.Storage
}

// NewValidator creates a validator with the given configuration.
func NewValidator(config Config) (Validator, error) {
	transport := otelhttp.NewTransport(http.DefaultTransport)
	client := &http.Client{
		Transport: transport,
	}

	jwksURL, err := url.Parse(config.JWKSURI)
	if err != nil {
		return nil, err
	}

	storageOpts := jwkset.HTTPClientStorageOptions{
		Client: client,
	}

	storage, err := jwkset.NewStorageFromHTTP(jwksURL, storageOpts)
	if err != nil {
		return nil, err
	}

	keyfuncOpts := keyfunc.Options{
		Storage: storage,
	}

	kf, err := keyfunc.New(keyfuncOpts)
	if err != nil {
		return nil, err
	}

	parser := jwt.NewParser(
		jwt.WithIssuedAt(),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(config.Issuer),
	)

	out := &validator{
		kf:     kf.Keyfunc,
		parser: parser,

		keyStorage: storage,
	}

	return out, nil
}

func (v *validator) ValidateToken(tokenString string) (string, map[string]any, error) {
	mapClaims := jwt.MapClaims{}

	_, err := v.parser.ParseWithClaims(tokenString, mapClaims, v.kf)
	if err != nil {
		return "", nil, err
	}

	sub, err := mapClaims.GetSubject()
	if err != nil {
		return "", nil, err
	}

	return sub, mapClaims, nil
}

// HealthCheck returns nil when the service is healthy.
func (v *validator) HealthCheck(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "HealthCheck")
	defer span.End()

	span.SetAttributes(attribute.String("healthcheck.outcome", "unhealthy"))

	keys, err := v.keyStorage.KeyReadAll(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	if len(keys) == 0 {
		span.SetStatus(codes.Error, ErrIssuerKeysMissing.Error())
		span.RecordError(ErrIssuerKeysMissing)

		return ErrIssuerKeysMissing
	}

	for _, key := range keys {
		if err := key.Validate(); err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return err
		}
	}

	span.SetAttributes(attribute.String("healthcheck.outcome", "healthy"))

	return nil
}
