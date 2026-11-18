package memory_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/RRanar/xp-expense-tracker/internal/domain/expense"
	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
	"github.com/RRanar/xp-expense-tracker/internal/infrastructure/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	alice = mustUser("alice")
	bob   = mustUser("bob")
)

func mustUser(v string) shared.UserID {
	u, err := shared.NewUserID(v)
	if err != nil {
		panic(err)
	}
	return u
}

func mustDate(v string) shared.Date {
	d, err := shared.DateFromString(v)
	if err != nil {
		panic(err)
	}
	return d
}

func newExpense(t *testing.T, owner shared.UserID, category, date string) *expense.Expense {
	t.Helper()
	e, err := expense.NewExpense(expense.NewParams{
		Amount:    100,
		Currency:  "USD",
		Category:  category,
		Merchant:  "Shop Synthetic",
		Note:      "note",
		Date:      mustDate(date),
		Owner:     owner,
		CreatedAt: shared.NewCreatedAt(time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)),
	})
	require.NoError(t, err)
	return e
}

func TestSavedExpenseIsLoadedBackWithEveryFieldIncludingID(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	orig := newExpense(t, alice, "Food", "2026-03-01")
	require.NoError(t, repo.Save(ctx, orig))

	got, err := repo.FindByID(ctx, alice, orig.ID())
	require.NoError(t, err)
	assert.Equal(t, orig, got)
}

func TestLoadedExpenseIsACopyNotTheStoredInstance(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	orig := newExpense(t, alice, "Food", "2026-03-01")
	require.NoError(t, repo.Save(ctx, orig))

	got, _ := repo.FindByID(ctx, alice, orig.ID())
	note := "mutated"
	require.NoError(t, got.Update(expense.Patch{Note: &note}))

	again, _ := repo.FindByID(ctx, alice, orig.ID())
	assert.Equal(t, "note", again.Note())
}

func TestAnotherOwnersExpenseIsNotFound(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	e := newExpense(t, alice, "Food", "2026-03-01")
	require.NoError(t, repo.Save(ctx, e))

	_, err := repo.FindByID(ctx, bob, e.ID())
	assert.ErrorIs(t, err, expense.ErrNotFound)
	assert.ErrorIs(t, repo.Delete(ctx, bob, e.ID()), expense.ErrNotFound)

	_, err = repo.FindByID(ctx, alice, e.ID())
	assert.NoError(t, err, "delete by another owner must not remove it")
}

func TestUnknownIDIsNotFound(t *testing.T) {
	repo := memory.NewRepository()
	_, err := repo.FindByID(context.Background(), alice, shared.NewID())
	assert.ErrorIs(t, err, expense.ErrNotFound)
	assert.ErrorIs(t, repo.Delete(context.Background(), alice, shared.NewID()), expense.ErrNotFound)
}

func TestUpdatePersistsChangesVisibleThroughFindByID(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	e := newExpense(t, alice, "Food", "2026-03-01")
	require.NoError(t, repo.Save(ctx, e))

	note := "changed"
	require.NoError(t, e.Update(expense.Patch{Note: &note}))
	require.NoError(t, repo.Update(ctx, e))

	got, err := repo.FindByID(ctx, alice, e.ID())
	require.NoError(t, err)
	assert.Equal(t, "changed", got.Note())
}

func TestUpdateOfUnknownOrForeignExpenseIsNotFound(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	mine := newExpense(t, alice, "Food", "2026-03-01")
	require.NoError(t, repo.Save(ctx, mine))

	assert.ErrorIs(t, repo.Update(ctx, newExpense(t, alice, "Food", "2026-03-01")), expense.ErrNotFound)

	foreign, err := expense.Reconstitute(expense.ReconstituteParams{
		ID: mine.ID(),
		NewParams: expense.NewParams{
			Amount: 1, Currency: "USD", Category: "Food",
			Date: mustDate("2026-03-01"), Owner: bob, CreatedAt: mine.CreatedAt(),
		},
	})
	require.NoError(t, err)
	assert.ErrorIs(t, repo.Update(ctx, foreign), expense.ErrNotFound)
}

func TestDeleteRemovesExpense(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	e := newExpense(t, alice, "Food", "2026-03-01")
	require.NoError(t, repo.Save(ctx, e))

	require.NoError(t, repo.Delete(ctx, alice, e.ID()))
	_, err := repo.FindByID(ctx, alice, e.ID())
	assert.ErrorIs(t, err, expense.ErrNotFound)
}

func TestEveryMethodReturnsTheContextErrorWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := memory.NewRepository()
	e := newExpense(t, alice, "Food", "2026-03-01")

	assert.ErrorIs(t, repo.Save(ctx, e), context.Canceled)
	_, err := repo.FindByID(ctx, alice, e.ID())
	assert.ErrorIs(t, err, context.Canceled)
	_, _, err = repo.List(ctx, alice, expense.Filter{}, expense.Page{})
	assert.ErrorIs(t, err, context.Canceled)
	assert.ErrorIs(t, repo.Update(ctx, e), context.Canceled)
	assert.ErrorIs(t, repo.Delete(ctx, alice, e.ID()), context.Canceled)
}

func TestListReturnsOnlyTheCallersExpenses(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	require.NoError(t, repo.Save(ctx, newExpense(t, alice, "Food", "2026-03-01")))
	require.NoError(t, repo.Save(ctx, newExpense(t, bob, "Food", "2026-03-02")))

	items, next, err := repo.List(ctx, alice, expense.Filter{}, expense.Page{})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, alice, items[0].Owner())
	assert.Empty(t, next)
}

func TestListFiltersByCategoryAndInclusiveDateRange(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	for _, x := range []struct{ cat, date string }{
		{"Food", "2026-03-01"}, {"Food", "2026-03-05"}, {"Food", "2026-03-10"},
		{"Transport", "2026-03-05"},
	} {
		require.NoError(t, repo.Save(ctx, newExpense(t, alice, x.cat, x.date)))
	}

	items, _, err := repo.List(ctx, alice, expense.Filter{Category: "Food"}, expense.Page{})
	require.NoError(t, err)
	assert.Len(t, items, 3)

	items, _, err = repo.List(ctx, alice, expense.Filter{
		Category: "Food", From: mustDate("2026-03-01"), To: mustDate("2026-03-05"),
	}, expense.Page{})
	require.NoError(t, err)
	require.Len(t, items, 2, "both bounds are inclusive")

	items, _, err = repo.List(ctx, alice, expense.Filter{From: mustDate("2026-03-06")}, expense.Page{})
	require.NoError(t, err)
	assert.Len(t, items, 1, "open-ended range")
}

func TestListIsOrderedByExpenseDateThenID(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	for _, d := range []string{"2026-03-09", "2026-03-01", "2026-03-05", "2026-03-05", "2026-03-05"} {
		require.NoError(t, repo.Save(ctx, newExpense(t, alice, "Food", d)))
	}

	items, _, err := repo.List(ctx, alice, expense.Filter{}, expense.Page{})
	require.NoError(t, err)
	require.Len(t, items, 5)
	for i := 1; i < len(items); i++ {
		prev, cur := items[i-1], items[i]
		c := prev.Date().Compare(cur.Date())
		assert.LessOrEqual(t, c, 0)
		if c == 0 {
			assert.Less(t, prev.ID().String(), cur.ID().String())
		}
	}
}

func TestListPagesWithOpaqueTokenThatResumesAfterTheLastItem(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Save(ctx, newExpense(t, alice, "Food", fmt.Sprintf("2026-03-0%d", i+1))))
	}

	var seen []string
	token := ""
	pages := 0
	for {
		items, next, err := repo.List(ctx, alice, expense.Filter{}, expense.Page{Size: 2, Token: token})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(items), 2)
		for _, e := range items {
			seen = append(seen, e.Date().String())
		}
		pages++
		if next == "" {
			break
		}
		token = next
	}
	assert.Equal(t, 3, pages)
	assert.Equal(t, []string{"2026-03-01", "2026-03-02", "2026-03-03", "2026-03-04", "2026-03-05"}, seen)
}

func TestListWithExactlyOnePageReturnsNoToken(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	for i := 1; i <= 2; i++ {
		require.NoError(t, repo.Save(ctx, newExpense(t, alice, "Food", fmt.Sprintf("2026-03-0%d", i))))
	}
	items, next, err := repo.List(ctx, alice, expense.Filter{}, expense.Page{Size: 2})
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Empty(t, next)
}

func TestListRejectsInvalidToken(t *testing.T) {
	_, _, err := memory.NewRepository().List(context.Background(), alice, expense.Filter{}, expense.Page{Token: "%%%"})
	assert.ErrorIs(t, err, shared.ErrInvalidInput)
}
