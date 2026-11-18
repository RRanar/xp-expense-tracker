package usecase_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/RRanar/xp-expense-tracker/internal/application/usecase"
	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
	"github.com/RRanar/xp-expense-tracker/internal/infrastructure/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seed(t *testing.T, repo *memory.Repository, owner, category string, n int) {
	t.Helper()
	create := usecase.NewCreateExpenseUseCase(repo, fixedClock{at: instant})
	for i := 0; i < n; i++ {
		in := validCreate()
		in.Owner = owner
		in.Category = category
		in.Date = fmt.Sprintf("2026-%02d-%02d", 1+i/28%12, 1+i%28)
		_, err := create.Execute(context.Background(), in)
		require.NoError(t, err)
	}
}

func TestListExpensesReturnsOnlyTheCallersExpenses(t *testing.T) {
	repo := memory.NewRepository()
	seed(t, repo, "alice", "Food", 2)
	seed(t, repo, "bob", "Transport", 3)

	out, err := usecase.NewListExpensesUseCase(repo).Execute(context.Background(), usecase.ListExpensesInput{Owner: "alice"})
	require.NoError(t, err)
	assert.Len(t, out.Items, 2)
	for _, it := range out.Items {
		assert.Equal(t, "Food", it.Category)
	}
}

func TestListExpensesFiltersByCategoryAndDateRange(t *testing.T) {
	repo := memory.NewRepository()
	seed(t, repo, "alice", "Food", 10)
	seed(t, repo, "alice", "Transport", 10)

	out, err := usecase.NewListExpensesUseCase(repo).Execute(context.Background(), usecase.ListExpensesInput{
		Owner: "alice", Category: "Food", From: "2026-01-03", To: "2026-01-05",
	})
	require.NoError(t, err)
	assert.Len(t, out.Items, 3)
}

func TestListExpensesDefaultsTo50AndCapsAt200(t *testing.T) {
	repo := memory.NewRepository()
	seed(t, repo, "alice", "Food", 205)
	uc := usecase.NewListExpensesUseCase(repo)

	out, err := uc.Execute(context.Background(), usecase.ListExpensesInput{Owner: "alice"})
	require.NoError(t, err)
	assert.Len(t, out.Items, 50)
	assert.NotEmpty(t, out.NextPageToken)

	out, err = uc.Execute(context.Background(), usecase.ListExpensesInput{Owner: "alice", PageSize: 1000})
	require.NoError(t, err)
	assert.Len(t, out.Items, 200)
	assert.NotEmpty(t, out.NextPageToken)
}

func TestListExpensesTokenResumesAfterTheLastItem(t *testing.T) {
	repo := memory.NewRepository()
	seed(t, repo, "alice", "Food", 5)
	uc := usecase.NewListExpensesUseCase(repo)

	first, err := uc.Execute(context.Background(), usecase.ListExpensesInput{Owner: "alice", PageSize: 3})
	require.NoError(t, err)
	require.Len(t, first.Items, 3)

	rest, err := uc.Execute(context.Background(), usecase.ListExpensesInput{Owner: "alice", PageSize: 3, PageToken: first.NextPageToken})
	require.NoError(t, err)
	assert.Len(t, rest.Items, 2)
	assert.Empty(t, rest.NextPageToken)
	assert.Greater(t, rest.Items[0].Date, first.Items[2].Date)
}

func TestListExpensesRejectsBadFilterAndToken(t *testing.T) {
	uc := usecase.NewListExpensesUseCase(memory.NewRepository())
	for _, in := range []usecase.ListExpensesInput{
		{Owner: "alice", From: "soon"},
		{Owner: "alice", To: "later"},
		{Owner: "alice", PageToken: "%%%"},
		{Owner: ""},
	} {
		_, err := uc.Execute(context.Background(), in)
		assert.ErrorIs(t, err, shared.ErrInvalidInput, "%+v", in)
	}
}

func TestListExpensesPassesTheContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := usecase.NewListExpensesUseCase(memory.NewRepository()).Execute(ctx, usecase.ListExpensesInput{Owner: "alice"})
	assert.ErrorIs(t, err, context.Canceled)
}
