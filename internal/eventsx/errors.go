package eventsx

import "errors"

var (
	// ErrPublishNotEnabled represents an error state where an event publish was attempted despite not being enabled
	ErrPublishNotEnabled = errors.New("event publishing is not enabled")

	// ErrPublisherNotConnected is returned when the underlying connection status is not CONNECTED.
	ErrPublisherNotConnected = errors.New("event publisher is not connected")
)
