package permissions

import "errors"

var (
	// ErrUnauthenticated represents an error state where the subject failed to authenticate
	// against permissions-api.
	ErrUnauthenticated = errors.New("invalid credentials")

	// ErrPermissionDenied represents an error state where the subject was denied access to
	// perform some action on a resource.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrUnexpectedResponse represents an error state where permissions-api returned an
	// unexpected response.
	ErrUnexpectedResponse = errors.New("unexpected response from server")
)
