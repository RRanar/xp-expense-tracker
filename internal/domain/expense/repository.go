package expense

import (
	"context"

	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
)

// Every method is scoped to an owner: a record of another owner is reported
// as ErrNotFound, never as forbidden, so existence is not leaked.
type Repository interface {
	Save(ctx context.Context, e *Expense) error
	FindByID(ctx context.Context, owner shared.UserID, id shared.ID) (*Expense, error)
	// List returns expenses ordered by expense date then id, and the token for
	// the next page, or an empty token when none remain.
	List(ctx context.Context, owner shared.UserID, f Filter, p Page) ([]*Expense, string, error)
	Update(ctx context.Context, e *Expense) error
	Delete(ctx context.Context, owner shared.UserID, id shared.ID) error
}
