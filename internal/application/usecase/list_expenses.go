package usecase

import (
	"context"

	"github.com/RRanar/xp-expense-tracker/internal/domain/expense"
	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
)

type ListExpensesUseCase struct {
	repo expense.Repository
}

type ListExpensesInput struct {
	Owner     string
	Category  string
	From      string
	To        string
	PageSize  int
	PageToken string
}

type ListExpensesOutput struct {
	Items         []ExpenseView
	NextPageToken string
}

func NewListExpensesUseCase(repo expense.Repository) *ListExpensesUseCase {
	return &ListExpensesUseCase{repo: repo}
}

func (uc *ListExpensesUseCase) Execute(ctx context.Context, in ListExpensesInput) (*ListExpensesOutput, error) {
	owner, err := shared.NewUserID(in.Owner)
	if err != nil {
		return nil, err
	}
	filter := expense.Filter{Category: in.Category}
	if in.From != "" {
		if filter.From, err = shared.DateFromString(in.From); err != nil {
			return nil, err
		}
	}
	if in.To != "" {
		if filter.To, err = shared.DateFromString(in.To); err != nil {
			return nil, err
		}
	}

	items, next, err := uc.repo.List(ctx, owner, filter, expense.Page{Size: in.PageSize, Token: in.PageToken})
	if err != nil {
		return nil, err
	}

	out := &ListExpensesOutput{Items: make([]ExpenseView, 0, len(items)), NextPageToken: next}
	for _, e := range items {
		out.Items = append(out.Items, viewOf(e))
	}
	return out, nil
}
