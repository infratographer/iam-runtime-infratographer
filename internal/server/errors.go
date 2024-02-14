package server

import "errors"

var (
	// ErrDuplicateValue represents an error where a duplicate value was found in a policy.
	ErrDuplicateValue = errors.New("duplicate value")
	// ErrMissingValue represents an error where a required value was missing from a policy.
	ErrMissingValue = errors.New("missing value")
)
