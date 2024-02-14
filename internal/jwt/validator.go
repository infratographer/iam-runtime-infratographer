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
	ValidateToken(string) (map[string]string, error)
}

type validator struct {
	kf jwt.Keyfunc
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

	out := &validator{
		kf: kf.Keyfunc,
	}

	return out, nil
}

func (v *validator) ValidateToken(tokenString string) (map[string]string, error) {
	tok, err := jwt.Parse(tokenString, v.kf)
	if err != nil {
		return nil, err
	}

	iss, err := tok.Claims.GetIssuer()
	if err != nil {
		return nil, err
	}

	sub, err := tok.Claims.GetSubject()
	if err != nil {
		return nil, err
	}

	out := map[string]string{
		"iss": iss,
		"sub": sub,
	}

	return out, nil
}
