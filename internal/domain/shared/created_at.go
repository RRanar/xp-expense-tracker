// Package shared holds the value objects and ports used across the domain:
// identifiers, money, dates, the clock and the base input error.
package shared

import (
	"fmt"
	"time"
)

type CreatedAt struct {
	value time.Time
}

func NewCreatedAt(t time.Time) CreatedAt {
	return CreatedAt{value: t}
}

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
