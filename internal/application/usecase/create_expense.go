// Package usecase holds the application use cases: one type per command or
// query. They orchestrate the domain and speak in plain strings and numbers so
// transport adapters stay free of domain types.
package usecase

import (
	"context"

	"github.com/RRanar/xp-expense-tracker/internal/domain/expense"
	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
)

type ExpenseView struct {
	ID        string
	Owner     string
	Amount    int64
	Currency  string
	Category  string
	Merchant  string
	Note      string
	Date      string
	CreatedAt string
}

func viewOf(e *expense.Expense) ExpenseView {
	return ExpenseView{
		ID:        e.ID().String(),
		Owner:     e.Owner().String(),
		Amount:    e.Amount(),
		Currency:  e.Currency().String(),
		Category:  e.Category(),
		Merchant:  e.Merchant(),
		Note:      e.Note(),
		Date:      e.Date().String(),
		CreatedAt: e.CreatedAt().String(),
	}
}

type CreateExpenseUseCase struct {
	repo  expense.Repository
	clock shared.Clock
}

type CreateExpenseInput struct {
	Owner    string
	Amount   int64
	Currency string
	Category string
	Merchant string
	Note     string
	Date     string
}

func NewCreateExpenseUseCase(repo expense.Repository, clock shared.Clock) *CreateExpenseUseCase {
	return &CreateExpenseUseCase{repo: repo, clock: clock}
}

func (uc *CreateExpenseUseCase) Execute(ctx context.Context, in CreateExpenseInput) (*ExpenseView, error) {
	owner, err := shared.NewUserID(in.Owner)
	if err != nil {
		return nil, err
	}
	date, err := shared.DateFromString(in.Date)
	if err != nil {
		return nil, err
	}

	e, err := expense.NewExpense(expense.NewParams{
		Amount:    in.Amount,
		Currency:  in.Currency,
		Category:  in.Category,
		Merchant:  in.Merchant,
		Note:      in.Note,
		Date:      date,
		Owner:     owner,
		CreatedAt: shared.NewCreatedAt(uc.clock.Now()),
	})
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, e); err != nil {
		return nil, err
	}

	view := viewOf(e)
	return &view, nil
}
