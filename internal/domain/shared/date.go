package shared

import (
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

type Date struct {
	value time.Time
}

func DateFromString(v string) (Date, error) {
	t, err := time.Parse(dateLayout, v)
	if err != nil {
		return Date{}, fmt.Errorf("%w: date %q must be YYYY-MM-DD", ErrInvalidInput, v)
	}
	return Date{value: t}, nil
}

func (d Date) Compare(other Date) int { return d.value.Compare(other.value) }
func (d Date) IsZero() bool           { return d.value.IsZero() }
func (d Date) String() string         { return d.value.Format(dateLayout) }
