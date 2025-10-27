package usecase

import (
	expense "github.com/RRanar/xp-expense-tracker/internal/domain/expence"
)

type ListExpensesUseCase struct {
	repo expense.Repository
}

type ListExpenseOutput struct {
	Amount      int64  `json:"amount"`
	Category    string `json:"category"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
}

func NewListExpensesUseCase(repo expense.Repository) *ListExpensesUseCase {
	return &ListExpensesUseCase{repo: repo}
}

func (uc *ListExpensesUseCase) Execute() ([]ListExpenseOutput, error) {
	items, err := uc.repo.FindAll()
	if err != nil {
		return nil, err
	}

	var out []ListExpenseOutput
	for _, e := range items {
		out = append(out, ListExpenseOutput{
			Amount:      e.Amount,
			Category:    e.Category,
			Description: e.Description,
			CreatedAt:   e.CreatedAt.Get(),
		})
	}

	return out, nil
}
