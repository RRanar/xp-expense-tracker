package shared_test

import (
	"testing"

	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDateParsesCalendarDay(t *testing.T) {
	d, err := shared.DateFromString("2026-02-28")
	require.NoError(t, err)
	assert.Equal(t, "2026-02-28", d.String())
	assert.False(t, d.IsZero())
}

func TestDateRejectsInvalidDays(t *testing.T) {
	for _, bad := range []string{"", "2026-02-30", "2026-13-01", "26-01-01", "2026-01-01T00:00:00Z", "today"} {
		_, err := shared.DateFromString(bad)
		assert.ErrorIs(t, err, shared.ErrInvalidInput, bad)
	}
}

func TestDateCompareOrdersChronologically(t *testing.T) {
	a, _ := shared.DateFromString("2026-01-31")
	b, _ := shared.DateFromString("2026-02-01")
	assert.Equal(t, -1, a.Compare(b))
	assert.Equal(t, 1, b.Compare(a))
	assert.Equal(t, 0, a.Compare(a))
}

func TestZeroDateIsZero(t *testing.T) {
	assert.True(t, shared.Date{}.IsZero())
}
