package eventsx

import "errors"

// ErrPublishNotEnabled represents an error state where an event publish was attempted despite not being enabled
var ErrPublishNotEnabled = errors.New("event publishing is not enabled")
