package expense

import (
	"strings"

	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
)

type Expense struct {
	id          shared.ID
	amount      int64
	category    string
	description string
	createdAt   shared.CreatedAt
}

func NewExpense(amount int64, category string, description string, createdAt string) (*Expense, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	trimmed := strings.TrimSpace(category)
	if trimmed == "" {
		return nil, ErrMissingCategory
	}

	return &Expense{
		id:          shared.NewID(),
		amount:      amount,
		category:    trimmed,
		description: description,
		createdAt:   shared.CreatedAtFromString(createdAt),
	}, nil
}

func (e *Expense) ID() shared.ID               { return e.id }
func (e *Expense) Amount() int64               { return e.amount }
func (e *Expense) Category() string            { return e.category }
func (e *Expense) Description() string         { return e.description }
func (e *Expense) CreatedAt() shared.CreatedAt { return e.createdAt }
