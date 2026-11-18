package usecase

import (
	"context"
	"fmt"

	"github.com/RRanar/xp-expense-tracker/internal/domain/expense"
	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
)

type GetExpenseInput struct {
	Owner string
	ID    string
}

type GetExpenseUseCase struct {
	repo expense.Repository
}

func NewGetExpenseUseCase(repo expense.Repository) *GetExpenseUseCase {
	return &GetExpenseUseCase{repo: repo}
}

func (uc *GetExpenseUseCase) Execute(ctx context.Context, in GetExpenseInput) (*ExpenseView, error) {
	owner, err := shared.NewUserID(in.Owner)
	if err != nil {
		return nil, err
	}
	e, err := uc.repo.FindByID(ctx, owner, shared.IDFromString(in.ID))
	if err != nil {
		return nil, err
	}
	view := viewOf(e)
	return &view, nil
}

type DeleteExpenseInput struct {
	Owner string
	ID    string
}

type DeleteExpenseUseCase struct {
	repo expense.Repository
}

func NewDeleteExpenseUseCase(repo expense.Repository) *DeleteExpenseUseCase {
	return &DeleteExpenseUseCase{repo: repo}
}

func (uc *DeleteExpenseUseCase) Execute(ctx context.Context, in DeleteExpenseInput) error {
	owner, err := shared.NewUserID(in.Owner)
	if err != nil {
		return err
	}
	return uc.repo.Delete(ctx, owner, shared.IDFromString(in.ID))
}

type ExpenseValues struct {
	Amount   int64
	Currency string
	Category string
	Merchant string
	Note     string
	Date     string
}

// Mask names the fields of Values to apply; the others are ignored.
type UpdateExpenseInput struct {
	Owner  string
	ID     string
	Mask   []string
	Values ExpenseValues
}

type UpdateExpenseUseCase struct {
	repo expense.Repository
}

func NewUpdateExpenseUseCase(repo expense.Repository) *UpdateExpenseUseCase {
	return &UpdateExpenseUseCase{repo: repo}
}

func (uc *UpdateExpenseUseCase) Execute(ctx context.Context, in UpdateExpenseInput) (*ExpenseView, error) {
	owner, err := shared.NewUserID(in.Owner)
	if err != nil {
		return nil, err
	}
	patch, err := patchFromMask(in.Mask, in.Values)
	if err != nil {
		return nil, err
	}

	e, err := uc.repo.FindByID(ctx, owner, shared.IDFromString(in.ID))
	if err != nil {
		return nil, err
	}
	if err := e.Update(patch); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, err
	}

	view := viewOf(e)
	return &view, nil
}

func patchFromMask(mask []string, v ExpenseValues) (expense.Patch, error) {
	if len(mask) == 0 {
		return expense.Patch{}, fmt.Errorf("%w: update mask is empty", shared.ErrInvalidInput)
	}
	var p expense.Patch
	for _, field := range mask {
		switch field {
		case "amount":
			p.Amount = &v.Amount
		case "currency":
			p.Currency = &v.Currency
		case "category":
			p.Category = &v.Category
		case "merchant":
			p.Merchant = &v.Merchant
		case "note":
			p.Note = &v.Note
		case "expense_date":
			date, err := shared.DateFromString(v.Date)
			if err != nil {
				return expense.Patch{}, err
			}
			p.Date = &date
		default:
			return expense.Patch{}, fmt.Errorf("%w: field %q cannot be updated", shared.ErrInvalidInput, field)
		}
	}
	return p, nil
}
