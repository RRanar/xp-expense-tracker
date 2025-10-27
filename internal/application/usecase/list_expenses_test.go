package usecase_test

import (
	"testing"

	"github.com/RRanar/xp-expense-tracker/internal/application/usecase"
	expense "github.com/RRanar/xp-expense-tracker/internal/domain/expence"
	"github.com/stretchr/testify/assert"
)

type fakeRepoList struct {
	items []*expense.Expense
}

func (f *fakeRepoList) Save(e *expense.Expense) error {
	f.items = append(f.items, e)

	return nil
}

func (f *fakeRepoList) FindAll() ([]*expense.Expense, error) {
	return f.items, nil
}

func TestListExpensesReturnAllSavedExpenses(t *testing.T) {
	repo := &fakeRepoList{}
	el1, _ := expense.NewExpense(1000, "Food", "Pizza")
	el2, _ := expense.NewExpense(2000, "Transport", "Taxi")
	repo.Save(el1)
	repo.Save(el2)

	uc := usecase.NewListExpensesUseCase(repo)
	out, err := uc.Execute()
	assert.NoError(t, err)
	assert.Len(t, out, 2)
	assert.Equal(t, "Food", out[0].Category)
	assert.Equal(t, "Transport", out[1].Category)
}
