// Package shared consists the VO and other object that helps to build buisness logic
package shared

import (
	"fmt"
	"time"
)

// CreatedAt is the instant an aggregate was created. It is always supplied
// by a Clock or loaded from storage; it never reads the system time itself.
type CreatedAt struct {
	value time.Time
}

func NewCreatedAt(t time.Time) CreatedAt {
	return CreatedAt{value: t}
}

// CreatedAtFromString parses an RFC 3339 instant. Bad input is an error;
// no substitute value is ever returned.
func CreatedAtFromString(v string) (CreatedAt, error) {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return CreatedAt{}, fmt.Errorf("%w: created at %q is not RFC 3339", ErrInvalidInput, v)
	}
	return CreatedAt{value: t}, nil
}

func (c CreatedAt) Time() time.Time { return c.value }
func (c CreatedAt) IsZero() bool    { return c.value.IsZero() }

func (c CreatedAt) String() string {
	return c.value.Format(time.RFC3339)
}
