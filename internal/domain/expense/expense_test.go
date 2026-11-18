package expense_test

import (
	"testing"
	"time"

	"github.com/RRanar/xp-expense-tracker/internal/domain/expense"
	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixedNow = time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)

func validParams() expense.NewParams {
	d, _ := shared.DateFromString("2026-03-01")
	owner, _ := shared.NewUserID("user-1")
	return expense.NewParams{
		Amount:    2500,
		Currency:  "USD",
		Category:  "Food",
		Merchant:  "Cafe Synthetic",
		Note:      "Lunch",
		Date:      d,
		Owner:     owner,
		CreatedAt: shared.NewCreatedAt(fixedNow),
	}
}

func TestExpenseCreatedSuccessfully(t *testing.T) {
	p := validParams()
	exp, err := expense.NewExpense(p)
	require.NoError(t, err)

	assert.Equal(t, int64(2500), exp.Amount())
	assert.Equal(t, "USD", exp.Currency().String())
	assert.Equal(t, "Food", exp.Category())
	assert.Equal(t, "Cafe Synthetic", exp.Merchant())
	assert.Equal(t, "Lunch", exp.Note())
	assert.Equal(t, "2026-03-01", exp.Date().String())
	assert.Equal(t, "user-1", exp.Owner().String())
	assert.Equal(t, fixedNow, exp.CreatedAt().Time())
	assert.NotEmpty(t, exp.ID().String())
}

func TestExpenseRejectsBrokenInvariants(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*expense.NewParams)
		want   error
	}{
		{"zero amount", func(p *expense.NewParams) { p.Amount = 0 }, expense.ErrInvalidAmount},
		{"negative amount", func(p *expense.NewParams) { p.Amount = -100 }, expense.ErrInvalidAmount},
		{"empty category", func(p *expense.NewParams) { p.Category = "" }, expense.ErrMissingCategory},
		{"blank category", func(p *expense.NewParams) { p.Category = "   " }, expense.ErrMissingCategory},
		{"lowercase currency", func(p *expense.NewParams) { p.Currency = "usd" }, shared.ErrInvalidInput},
		{"short currency", func(p *expense.NewParams) { p.Currency = "US" }, shared.ErrInvalidInput},
		{"missing owner", func(p *expense.NewParams) { p.Owner = shared.UserID{} }, expense.ErrMissingOwner},
		{"missing date", func(p *expense.NewParams) { p.Date = shared.Date{} }, expense.ErrInvalidDate},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := validParams()
			tc.mutate(&p)
			exp, err := expense.NewExpense(p)
			assert.ErrorIs(t, err, tc.want)
			assert.Nil(t, exp)
		})
	}
}

func TestExpenseAmountErrorKeepsMessage(t *testing.T) {
	p := validParams()
	p.Amount = 0
	_, err := expense.NewExpense(p)
	assert.EqualError(t, err, "amount must be positive")
}

func TestReconstitutePreservesEveryFieldIncludingID(t *testing.T) {
	orig, err := expense.NewExpense(validParams())
	require.NoError(t, err)

	got, err := expense.Reconstitute(expense.ReconstituteParams{
		ID: orig.ID(),
		NewParams: expense.NewParams{
			Amount:    orig.Amount(),
			Currency:  orig.Currency().String(),
			Category:  orig.Category(),
			Merchant:  orig.Merchant(),
			Note:      orig.Note(),
			Date:      orig.Date(),
			Owner:     orig.Owner(),
			CreatedAt: orig.CreatedAt(),
		},
	})
	require.NoError(t, err)
	assert.Equal(t, orig, got)
}

func TestReconstituteRejectsMissingID(t *testing.T) {
	_, err := expense.Reconstitute(expense.ReconstituteParams{NewParams: validParams()})
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
}

func ptr[T any](v T) *T { return &v }

func TestUpdateChangesOnlyGivenFields(t *testing.T) {
	exp, err := expense.NewExpense(validParams())
	require.NoError(t, err)

	err = exp.Update(expense.Patch{Amount: ptr(int64(999)), Note: ptr("Dinner")})
	require.NoError(t, err)

	assert.Equal(t, int64(999), exp.Amount())
	assert.Equal(t, "Dinner", exp.Note())
	assert.Equal(t, "Food", exp.Category())
	assert.Equal(t, "Cafe Synthetic", exp.Merchant())
	assert.Equal(t, "USD", exp.Currency().String())
	assert.Equal(t, "2026-03-01", exp.Date().String())
}

func TestUpdateRechecksInvariantsAndLeavesExpenseUntouchedOnError(t *testing.T) {
	exp, err := expense.NewExpense(validParams())
	require.NoError(t, err)

	err = exp.Update(expense.Patch{Note: ptr("changed"), Amount: ptr(int64(0))})
	assert.ErrorIs(t, err, expense.ErrInvalidAmount)
	assert.Equal(t, "Lunch", exp.Note(), "failed update must not apply partially")
	assert.Equal(t, int64(2500), exp.Amount())

	err = exp.Update(expense.Patch{Category: ptr(" ")})
	assert.ErrorIs(t, err, expense.ErrMissingCategory)
	err = exp.Update(expense.Patch{Currency: ptr("xx")})
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
}
