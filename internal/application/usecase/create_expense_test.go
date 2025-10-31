package usecase_test

import (
	"testing"

	"github.com/RRanar/xp-expense-tracker/internal/application/usecase"
	"github.com/RRanar/xp-expense-tracker/internal/domain/expence"
	"github.com/stretchr/testify/assert"
)

type fakeRepo struct {
	saved []*expense.Expense
}

func (f *fakeRepo) Save(e *expense.Expense) error {
	f.saved = append(f.saved, e)

	return nil
}

func (f *fakeRepo) FindAll() ([]*expense.Expense, error) {
	return f.saved, nil
}

func TestCreateExpenseSuccess(t *testing.T) {
	repo := &fakeRepo{}
	uc := usecase.NewCreateExpenseUseCase(repo)

	input := usecase.CreateExpenseInput{
		Amount:      2500,
		Category:    "Food",
		Description: "Lunch and cafe",
	}

	output, err := uc.Execute(input)
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, 1, len(repo.saved))
	assert.Equal(t, "Food", repo.saved[0].Category())
}
