package shared

import (
	"fmt"
	"strings"
)

// UserID identifies the owner of a record. Where it comes from is decided
// in a later spec (authentication); the domain only requires it to be present.
type UserID struct {
	value string
}

func NewUserID(v string) (UserID, error) {
	if strings.TrimSpace(v) == "" {
		return UserID{}, fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}
	return UserID{value: v}, nil
}

func (u UserID) String() string { return u.value }
func (u UserID) IsZero() bool   { return u.value == "" }
