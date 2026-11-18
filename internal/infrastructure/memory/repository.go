// Package memory is an in-memory expense.Repository for tests. It stores flat
// records and rebuilds entities through Reconstitute, as a database adapter
// would. It makes no performance claim.
package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/RRanar/xp-expense-tracker/internal/domain/expense"
	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
)

type Repository struct {
	mu      sync.RWMutex
	records map[string]expense.ReconstituteParams
}

func NewRepository() *Repository {
	return &Repository{records: make(map[string]expense.ReconstituteParams)}
}

func (r *Repository) Save(ctx context.Context, e *expense.Expense) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[e.ID().String()] = toRecord(e)
	return nil
}

func (r *Repository) FindByID(ctx context.Context, owner shared.UserID, id shared.ID) (*expense.Expense, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.records[id.String()]
	if !ok || rec.Owner != owner {
		return nil, expense.ErrNotFound
	}
	return expense.Reconstitute(rec)
}

func (r *Repository) List(ctx context.Context, owner shared.UserID, f expense.Filter, p expense.Page) ([]*expense.Expense, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	var after *expense.Cursor
	if p.Token != "" {
		c, err := expense.DecodeCursor(p.Token)
		if err != nil {
			return nil, "", err
		}
		after = &c
	}

	r.mu.RLock()
	matches := make([]expense.ReconstituteParams, 0, len(r.records))
	for _, rec := range r.records {
		if matchesFilter(rec, owner, f, after) {
			matches = append(matches, rec)
		}
	}
	r.mu.RUnlock()

	sort.Slice(matches, func(i, j int) bool {
		if c := matches[i].Date.Compare(matches[j].Date); c != 0 {
			return c < 0
		}
		return matches[i].ID.String() < matches[j].ID.String()
	})

	size := p.EffectiveSize()
	next := ""
	if len(matches) > size {
		matches = matches[:size]
		last := matches[size-1]
		next = expense.EncodeCursor(expense.Cursor{Date: last.Date, ID: last.ID})
	}

	out := make([]*expense.Expense, 0, len(matches))
	for _, rec := range matches {
		e, err := expense.Reconstitute(rec)
		if err != nil {
			return nil, "", err
		}
		out = append(out, e)
	}
	return out, next, nil
}

func (r *Repository) Update(ctx context.Context, e *expense.Expense) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.records[e.ID().String()]
	if !ok || existing.Owner != e.Owner() {
		return expense.ErrNotFound
	}
	r.records[e.ID().String()] = toRecord(e)
	return nil
}

func (r *Repository) Delete(ctx context.Context, owner shared.UserID, id shared.ID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.records[id.String()]
	if !ok || rec.Owner != owner {
		return expense.ErrNotFound
	}
	delete(r.records, id.String())
	return nil
}

func matchesFilter(rec expense.ReconstituteParams, owner shared.UserID, f expense.Filter, after *expense.Cursor) bool {
	if rec.Owner != owner {
		return false
	}
	if f.Category != "" && rec.Category != f.Category {
		return false
	}
	if !f.From.IsZero() && rec.Date.Compare(f.From) < 0 {
		return false
	}
	if !f.To.IsZero() && rec.Date.Compare(f.To) > 0 {
		return false
	}
	if after != nil {
		c := rec.Date.Compare(after.Date)
		if c < 0 || (c == 0 && rec.ID.String() <= after.ID.String()) {
			return false
		}
	}
	return true
}

func toRecord(e *expense.Expense) expense.ReconstituteParams {
	return expense.ReconstituteParams{
		ID: e.ID(),
		NewParams: expense.NewParams{
			Amount:    e.Amount(),
			Currency:  e.Currency().String(),
			Category:  e.Category(),
			Merchant:  e.Merchant(),
			Note:      e.Note(),
			Date:      e.Date(),
			Owner:     e.Owner(),
			CreatedAt: e.CreatedAt(),
		},
	}
}
