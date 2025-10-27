package expense

import (
	"errors"
	"strings"
	"time"
)

type Expense struct {
	Amount      int64
	Category    string
	Description string
	CreatedAt   *time.Time
}

func NewExpense(amount int64, category string, description string, createdAt *time.Time) (*Expense, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	trimmed := strings.TrimSpace(category)
	if trimmed == "" {
		return nil, errors.New("category is required")
	}

	var createdTime *time.Time

	if createdAt == nil {
		createdTime = &time.Time{}
	} else {
		createdTime = createdAt
	}

	return &Expense{
		Amount:      amount,
		Category:    trimmed,
		Description: description,
		CreatedAt:   createdTime,
	}, nil
}
