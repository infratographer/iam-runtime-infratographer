package filetokensource

import "errors"

// ErrTokenPathRequired is returned when the Config.TokenPath is not configured.
var ErrTokenPathRequired = errors.New("file token source: TokenPath required")

// Config describes the configuration for the token source.
type Config struct {
	// TokenPath is the path to the source jwt token.
	TokenPath string
}

// WithTokenPath returns a new Config with the provided token path defined.
func (c Config) WithTokenPath(path string) Config {
	c.TokenPath = path

	return c
}

// Configured returns true when TokenPath is defined.
func (c Config) Configured() bool {
	return c.TokenPath != ""
}

// Validate ensures the config has been configured properly.
func (c Config) Validate() error {
	if c.TokenPath == "" {
		return ErrTokenPathRequired
	}

	return nil
}

// ToTokenSource initializes a new [TokenSource] with the defined config.
func (c Config) ToTokenSource() (*TokenSource, error) {
	if c.TokenPath == "" {
		return nil, ErrTokenPathRequired
	}

	tokenSource := &TokenSource{
		path: c.TokenPath,
	}

	if _, err := tokenSource.Token(); err != nil {
		return nil, err
	}

	return tokenSource, nil
}
