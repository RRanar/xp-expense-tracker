package expense

import (
	"errors"
	"strings"
)

type Expense struct {
	Amount      int64
	Category    string
	Description string
}

func NewExpense(amount int64, category string, description string) (*Expense, error) {
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
	}, nil
}
