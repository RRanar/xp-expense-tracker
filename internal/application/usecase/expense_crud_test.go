package usecase_test

import (
	"context"
	"testing"

	"github.com/RRanar/xp-expense-tracker/internal/application/usecase"
	"github.com/RRanar/xp-expense-tracker/internal/domain/expense"
	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
	"github.com/RRanar/xp-expense-tracker/internal/infrastructure/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createOne(t *testing.T, repo *memory.Repository) string {
	t.Helper()
	out, err := usecase.NewCreateExpenseUseCase(repo, fixedClock{at: instant}).Execute(context.Background(), validCreate())
	require.NoError(t, err)
	return out.ID
}

func TestGetExpenseReturnsTheOwnersExpense(t *testing.T) {
	repo := memory.NewRepository()
	id := createOne(t, repo)

	out, err := usecase.NewGetExpenseUseCase(repo).Execute(context.Background(), usecase.GetExpenseInput{Owner: "alice", ID: id})
	require.NoError(t, err)
	assert.Equal(t, id, out.ID)
	assert.Equal(t, "Lunch", out.Note)
}

func TestGetExpenseOfAnotherOwnerIsNotFound(t *testing.T) {
	repo := memory.NewRepository()
	id := createOne(t, repo)

	_, err := usecase.NewGetExpenseUseCase(repo).Execute(context.Background(), usecase.GetExpenseInput{Owner: "bob", ID: id})
	assert.ErrorIs(t, err, expense.ErrNotFound)
}

func TestUpdateExpenseChangesOnlyMaskedFields(t *testing.T) {
	repo := memory.NewRepository()
	id := createOne(t, repo)

	out, err := usecase.NewUpdateExpenseUseCase(repo).Execute(context.Background(), usecase.UpdateExpenseInput{
		Owner: "alice",
		ID:    id,
		Mask:  []string{"note", "amount"},
		Values: usecase.ExpenseValues{
			Note: "Dinner", Amount: 999,
			Category: "ignored because not in mask",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Dinner", out.Note)
	assert.Equal(t, int64(999), out.Amount)
	assert.Equal(t, "Food", out.Category)

	got, err := repo.FindByID(context.Background(), mustOwner("alice"), shared.IDFromString(id))
	require.NoError(t, err)
	assert.Equal(t, "Dinner", got.Note())
	assert.Equal(t, "Food", got.Category())
}

func TestUpdateExpenseRechecksInvariants(t *testing.T) {
	repo := memory.NewRepository()
	id := createOne(t, repo)

	_, err := usecase.NewUpdateExpenseUseCase(repo).Execute(context.Background(), usecase.UpdateExpenseInput{
		Owner: "alice", ID: id, Mask: []string{"amount"}, Values: usecase.ExpenseValues{Amount: -1},
	})
	assert.ErrorIs(t, err, expense.ErrInvalidAmount)

	got, _ := repo.FindByID(context.Background(), mustOwner("alice"), shared.IDFromString(id))
	assert.Equal(t, int64(2500), got.Amount())
}

func TestUpdateExpenseRejectsEmptyOrUnknownMask(t *testing.T) {
	repo := memory.NewRepository()
	id := createOne(t, repo)
	uc := usecase.NewUpdateExpenseUseCase(repo)

	for _, mask := range [][]string{nil, {}, {"owner"}, {"note", "created_at"}} {
		_, err := uc.Execute(context.Background(), usecase.UpdateExpenseInput{Owner: "alice", ID: id, Mask: mask})
		assert.ErrorIs(t, err, shared.ErrInvalidInput, "%v", mask)
	}
}

func TestUpdateExpenseOfAnotherOwnerOrUnknownIDIsNotFound(t *testing.T) {
	repo := memory.NewRepository()
	id := createOne(t, repo)
	uc := usecase.NewUpdateExpenseUseCase(repo)
	values := usecase.ExpenseValues{Note: "x"}

	_, err := uc.Execute(context.Background(), usecase.UpdateExpenseInput{Owner: "bob", ID: id, Mask: []string{"note"}, Values: values})
	assert.ErrorIs(t, err, expense.ErrNotFound)
	_, err = uc.Execute(context.Background(), usecase.UpdateExpenseInput{Owner: "alice", ID: "missing", Mask: []string{"note"}, Values: values})
	assert.ErrorIs(t, err, expense.ErrNotFound)
}

func TestDeleteExpenseRemovesIt(t *testing.T) {
	repo := memory.NewRepository()
	id := createOne(t, repo)

	require.NoError(t, usecase.NewDeleteExpenseUseCase(repo).Execute(context.Background(), usecase.DeleteExpenseInput{Owner: "alice", ID: id}))
	_, err := usecase.NewGetExpenseUseCase(repo).Execute(context.Background(), usecase.GetExpenseInput{Owner: "alice", ID: id})
	assert.ErrorIs(t, err, expense.ErrNotFound)
}

func TestDeleteExpenseOfAnotherOwnerIsNotFoundAndKeepsIt(t *testing.T) {
	repo := memory.NewRepository()
	id := createOne(t, repo)

	err := usecase.NewDeleteExpenseUseCase(repo).Execute(context.Background(), usecase.DeleteExpenseInput{Owner: "bob", ID: id})
	assert.ErrorIs(t, err, expense.ErrNotFound)
	_, err = usecase.NewGetExpenseUseCase(repo).Execute(context.Background(), usecase.GetExpenseInput{Owner: "alice", ID: id})
	assert.NoError(t, err)
}

func TestGetUpdateDeletePassTheContext(t *testing.T) {
	repo := memory.NewRepository()
	id := createOne(t, repo)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := usecase.NewGetExpenseUseCase(repo).Execute(ctx, usecase.GetExpenseInput{Owner: "alice", ID: id})
	assert.ErrorIs(t, err, context.Canceled)
	_, err = usecase.NewUpdateExpenseUseCase(repo).Execute(ctx, usecase.UpdateExpenseInput{Owner: "alice", ID: id, Mask: []string{"note"}})
	assert.ErrorIs(t, err, context.Canceled)
	err = usecase.NewDeleteExpenseUseCase(repo).Execute(ctx, usecase.DeleteExpenseInput{Owner: "alice", ID: id})
	assert.ErrorIs(t, err, context.Canceled)
}
