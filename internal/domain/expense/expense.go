// Package expense is the Expense aggregate: its invariants, its errors and the
// repository port. It depends only on the shared domain package.
package expense

import (
	"fmt"
	"strings"

	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
)

type Expense struct {
	id        shared.ID
	money     shared.Money
	category  string
	merchant  string
	note      string
	date      shared.Date
	owner     shared.UserID
	createdAt shared.CreatedAt
}

type NewParams struct {
	Amount    int64
	Currency  string
	Category  string
	Merchant  string
	Note      string
	Date      shared.Date
	Owner     shared.UserID
	CreatedAt shared.CreatedAt
}

type ReconstituteParams struct {
	ID shared.ID
	NewParams
}

// Patch carries only the fields to change; nil means keep the current value.
type Patch struct {
	Amount   *int64
	Currency *string
	Category *string
	Merchant *string
	Note     *string
	Date     *shared.Date
}

func NewExpense(p NewParams) (*Expense, error) {
	return build(shared.NewID(), p)
}

// Reconstitute rebuilds an expense from storage keeping its id and created
// time, and re-checks the invariants so corrupt rows surface as errors.
func Reconstitute(p ReconstituteParams) (*Expense, error) {
	if p.ID.String() == "" {
		return nil, fmt.Errorf("%w: expense id is required", shared.ErrInvalidInput)
	}
	return build(p.ID, p.NewParams)
}

func build(id shared.ID, p NewParams) (*Expense, error) {
	if p.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	currency, err := shared.NewCurrency(p.Currency)
	if err != nil {
		return nil, err
	}
	category := strings.TrimSpace(p.Category)
	if category == "" {
		return nil, ErrMissingCategory
	}
	if p.Owner.IsZero() {
		return nil, ErrMissingOwner
	}
	if p.Date.IsZero() {
		return nil, ErrInvalidDate
	}

	return &Expense{
		id:        id,
		money:     shared.NewMoney(p.Amount, currency),
		category:  category,
		merchant:  p.Merchant,
		note:      p.Note,
		date:      p.Date,
		owner:     p.Owner,
		createdAt: p.CreatedAt,
	}, nil
}

// Update applies the patch atomically: on any invariant failure the expense is unchanged.
func (e *Expense) Update(p Patch) error {
	next := NewParams{
		Amount:    e.money.Minor(),
		Currency:  e.money.Currency().String(),
		Category:  e.category,
		Merchant:  e.merchant,
		Note:      e.note,
		Date:      e.date,
		Owner:     e.owner,
		CreatedAt: e.createdAt,
	}
	if p.Amount != nil {
		next.Amount = *p.Amount
	}
	if p.Currency != nil {
		next.Currency = *p.Currency
	}
	if p.Category != nil {
		next.Category = *p.Category
	}
	if p.Merchant != nil {
		next.Merchant = *p.Merchant
	}
	if p.Note != nil {
		next.Note = *p.Note
	}
	if p.Date != nil {
		next.Date = *p.Date
	}

	updated, err := build(e.id, next)
	if err != nil {
		return err
	}
	*e = *updated
	return nil
}

func (e *Expense) ID() shared.ID               { return e.id }
func (e *Expense) Amount() int64               { return e.money.Minor() }
func (e *Expense) Currency() shared.Currency   { return e.money.Currency() }
func (e *Expense) Category() string            { return e.category }
func (e *Expense) Merchant() string            { return e.merchant }
func (e *Expense) Note() string                { return e.note }
func (e *Expense) Date() shared.Date           { return e.date }
func (e *Expense) Owner() shared.UserID        { return e.owner }
func (e *Expense) CreatedAt() shared.CreatedAt { return e.createdAt }
