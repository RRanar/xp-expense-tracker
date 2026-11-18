package shared

import "time"

// Clock is the port for the current time. Use cases receive one so tests
// can fix the instant.
type Clock interface {
	Now() time.Time
}
