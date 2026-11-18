package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/RRanar/xp-expense-tracker/internal/application/usecase"
	"github.com/RRanar/xp-expense-tracker/internal/domain/expense"
	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
	"github.com/RRanar/xp-expense-tracker/internal/infrastructure/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fixedClock struct{ at time.Time }

func (c fixedClock) Now() time.Time { return c.at }

var instant = time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)

func validCreate() usecase.CreateExpenseInput {
	return usecase.CreateExpenseInput{
		Owner:    "alice",
		Amount:   2500,
		Currency: "USD",
		Category: "Food",
		Merchant: "Cafe Synthetic",
		Note:     "Lunch",
		Date:     "2026-03-01",
	}
}

func TestCreateExpenseUsesTheInjectedClockForCreatedAt(t *testing.T) {
	repo := memory.NewRepository()
	uc := usecase.NewCreateExpenseUseCase(repo, fixedClock{at: instant})

	out, err := uc.Execute(context.Background(), validCreate())
	require.NoError(t, err)
	assert.Equal(t, instant.Format(time.RFC3339), out.CreatedAt)

	stored, err := repo.FindByID(context.Background(), mustOwner("alice"), shared.IDFromString(out.ID))
	require.NoError(t, err)
	assert.Equal(t, instant, stored.CreatedAt().Time())
	assert.Equal(t, "Food", stored.Category())
	assert.Equal(t, "Cafe Synthetic", out.Merchant)
	assert.Equal(t, "2026-03-01", out.Date)
}

func TestCreateExpenseRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*usecase.CreateExpenseInput)
		want   error
	}{
		{"zero amount", func(i *usecase.CreateExpenseInput) { i.Amount = 0 }, expense.ErrInvalidAmount},
		{"missing category", func(i *usecase.CreateExpenseInput) { i.Category = "" }, expense.ErrMissingCategory},
		{"bad currency", func(i *usecase.CreateExpenseInput) { i.Currency = "usd" }, shared.ErrInvalidInput},
		{"missing owner", func(i *usecase.CreateExpenseInput) { i.Owner = "" }, shared.ErrInvalidInput},
		{"bad date", func(i *usecase.CreateExpenseInput) { i.Date = "tomorrow" }, shared.ErrInvalidInput},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := memory.NewRepository()
			in := validCreate()
			tc.mutate(&in)
			_, err := usecase.NewCreateExpenseUseCase(repo, fixedClock{at: instant}).Execute(context.Background(), in)
			assert.ErrorIs(t, err, tc.want)
		})
	}
}

func TestCreateExpensePassesTheContextToTheRepository(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	uc := usecase.NewCreateExpenseUseCase(memory.NewRepository(), fixedClock{at: instant})

	_, err := uc.Execute(ctx, validCreate())
	assert.ErrorIs(t, err, context.Canceled)
}

func mustOwner(v string) shared.UserID {
	u, err := shared.NewUserID(v)
	if err != nil {
		panic(err)
	}
	return u
}
