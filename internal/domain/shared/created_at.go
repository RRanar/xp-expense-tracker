// Package shared consists the VO and other object that helps to build buisness logic
package shared

import "time"

type CreatedAt struct {
	value time.Time
}

func NewCreatedAt() CreatedAt {
	return CreatedAt{value: time.Now()}
}

func CreatedAtFromString(v string) CreatedAt {
	if t, err := time.Parse(time.RFC3339, v); err != nil {
		return CreatedAt{value: time.Now()}
	} else {
		return CreatedAt{value: t}
	}
}

func (c CreatedAt) String() string {
	return c.value.Format(time.RFC3339)
}
