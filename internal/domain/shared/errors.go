package shared

import "errors"

// ErrInvalidInput marks a value that failed validation. Transport adapters
// map it to a status code at the edge.
var ErrInvalidInput = errors.New("invalid input")
