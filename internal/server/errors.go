package server

import "errors"

var (
	// ErrDuplicateValue represents an error where a duplicate value was found in a policy.
	ErrDuplicateValue = errors.New("duplicate value")
	// ErrMissingValue represents an error where a required value was missing from a policy.
	ErrMissingValue = errors.New("missing value")
	// ErrServerNotRunning is returned by the server health check when the server is not running.
	ErrServerNotRunning = errors.New("server not running")
)
