package expense

import "errors"

var (
	ErrInvalidAmount   = errors.New("amount must be positive")
	ErrMissingCategory = errors.New("category is required")
)
