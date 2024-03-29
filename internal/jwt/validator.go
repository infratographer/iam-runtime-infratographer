package jwt

import (
	"net/http"
	"net/url"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Validator represents a JWT validator.
type Validator interface {
	// ValidateToken checks that the given token is valid (i.e., is well-formed with a valid
	// signature and future expiry). On success, it returns a map of claims describing the subject.
	ValidateToken(string) (string, map[string]any, error)
}

type validator struct {
	kf     jwt.Keyfunc
	parser *jwt.Parser
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
