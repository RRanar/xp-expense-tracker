package expense

import "errors"

var (
	ErrInvalidAmount   = errors.New("amount must be positive")
	ErrMissingCategory = errors.New("category is required")
	ErrMissingOwner    = errors.New("owner is required")
	ErrInvalidDate     = errors.New("expense date is required")
	ErrNotFound        = errors.New("expense not found")
)
