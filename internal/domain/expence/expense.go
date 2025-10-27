package expense

import (
	"errors"
	"strings"
	"time"
)

type ExpenseCreatedTime struct {
	value time.Time
}

func NewExpenseCreatedTime(t *time.Time) *ExpenseCreatedTime {
	if t == nil {
		return &ExpenseCreatedTime{value: time.Now()}
	}

	return &ExpenseCreatedTime{value: *t}
}

func (ct *ExpenseCreatedTime) Get() string {
	return ct.value.Format(time.RFC3339)
}

type Expense struct {
	Amount      int64
	Category    string
	Description string
	CreatedAt   *ExpenseCreatedTime
}

func NewExpense(amount int64, category string, description string, createdAt *time.Time) (*Expense, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	trimmed := strings.TrimSpace(category)
	if trimmed == "" {
		return nil, errors.New("category is required")
	}

	return &Expense{
		Amount:      amount,
		Category:    trimmed,
		Description: description,
		CreatedAt:   NewExpenseCreatedTime(createdAt),
	}, nil
}
