// Package application layer that keeps all need parts of buisness logic orchestartion
package application

import "errors"

var (
	ErrValidation = errors.New("validation_error")
	ErrStorage    = errors.New("storage_error")
)
