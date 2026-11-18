package shared_test

import (
	"testing"
	"time"

	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatedAtFromStringRejectsInvalidInputInsteadOfUsingNow(t *testing.T) {
	for _, in := range []string{"", "yesterday", "2026-13-40T00:00:00Z", "2026-01-02"} {
		got, err := shared.CreatedAtFromString(in)
		assert.ErrorIs(t, err, shared.ErrInvalidInput, in)
		assert.True(t, got.IsZero(), "no substitute value for %q", in)
	}
}

func TestCreatedAtFromStringRoundTrips(t *testing.T) {
	got, err := shared.CreatedAtFromString("2026-03-04T05:06:07Z")
	require.NoError(t, err)
	assert.Equal(t, "2026-03-04T05:06:07Z", got.String())
}

func TestNewCreatedAtKeepsGivenInstant(t *testing.T) {
	at := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	assert.Equal(t, at, shared.NewCreatedAt(at).Time())
}
