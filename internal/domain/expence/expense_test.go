package expense_test

import (
	"testing"

	"github.com/RRanar/xp-expense-tracker/internal/domain/expence"
	"github.com/stretchr/testify/assert"
)

func TestExpenseCannotBeNegativeOrZero(t *testing.T) {
	_, err := expense.NewExpense(0, "Food", "Lunch")
	assert.Error(t, err)
	assert.EqualError(t, err, "amount must be positive")

	_, err = expense.NewExpense(-100, "Food", "Lunch")
	assert.Error(t, err)
	assert.EqualError(t, err, "amount must be positive")
}

func TestExpenseCannotBeWithoutCategory(t *testing.T) {
	_, err := expense.NewExpense(10, "", "Lunch")
	assert.Error(t, err)
	assert.EqualError(t, err, "category is required")

	_, err = expense.NewExpense(20, "   ", "Lunch")
	assert.Error(t, err)
	assert.EqualError(t, err, "category is required")
}

func TestExpenseCreatedSuccessfully(t *testing.T) {
	exp, err := expense.NewExpense(2500, "Food", "Lunch and cafe")
	assert.NoError(t, err)
	assert.NotNil(t, exp)

	assert.Equal(t, int64(2500), exp.Amount)
	assert.Equal(t, "Food", exp.Category)
	assert.Equal(t, "Lunch and cafe", exp.Description)
}
