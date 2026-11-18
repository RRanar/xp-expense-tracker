package expense_test

import (
	"testing"

	"github.com/RRanar/xp-expense-tracker/internal/domain/expense"
	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageSizeDefaultsAndCaps(t *testing.T) {
	tests := []struct{ in, want int }{
		{0, 50},
		{-5, 50},
		{1, 1},
		{50, 50},
		{200, 200},
		{201, 200},
		{100000, 200},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, expense.Page{Size: tc.in}.EffectiveSize(), "size %d", tc.in)
	}
}

func TestCursorRoundTripsAndIsOpaque(t *testing.T) {
	d, _ := shared.DateFromString("2026-03-01")
	id := shared.IDFromString("0b9f4c7e-0000-4000-8000-000000000001")

	token := expense.EncodeCursor(expense.Cursor{Date: d, ID: id})
	assert.NotContains(t, token, "2026-03-01")

	got, err := expense.DecodeCursor(token)
	require.NoError(t, err)
	assert.Equal(t, d, got.Date)
	assert.Equal(t, id, got.ID)
}

func TestDecodeCursorRejectsGarbage(t *testing.T) {
	for _, bad := range []string{"not base64 !!", "Zm9v", "e30"} {
		_, err := expense.DecodeCursor(bad)
		assert.ErrorIs(t, err, shared.ErrInvalidInput, bad)
	}
}
