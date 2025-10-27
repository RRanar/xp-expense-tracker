// Package usecase provides all UseCases for application layer
package usecase

import (
	"time"

	expense "github.com/RRanar/xp-expense-tracker/internal/domain/expence"
)

type CreateExpenseUseCase struct {
	repo expense.Repository
}

type CreateExpenseInput struct {
	Amount      int64
	Category    string
	Description string
}

type CreateExpenseOutput struct {
	ID          string `json:"id"`
	Amount      int64  `json:"amount"`
	Category    string `json:"category"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
}

func NewCreateExpenseUseCase(repo expense.Repository) *CreateExpenseUseCase {
	return &CreateExpenseUseCase{repo: repo}
}

func (uc *CreateExpenseUseCase) Execute(in CreateExpenseInput) (*CreateExpenseOutput, error) {
	createdAt := time.Now()
	e, err := expense.NewExpense(in.Amount, in.Category, in.Description, &createdAt)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(e); err != nil {
		return nil, err
	}

	out := &CreateExpenseOutput{
		Amount:      e.Amount,
		Category:    e.Category,
		Description: e.Description,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
	}

	return out, nil
}
