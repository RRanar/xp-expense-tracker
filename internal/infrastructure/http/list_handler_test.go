package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	app "github.com/RRanar/xp-expense-tracker/internal/application/usecase"
	expense "github.com/RRanar/xp-expense-tracker/internal/domain/expence"
	handler "github.com/RRanar/xp-expense-tracker/internal/infrastructure/http"
	"github.com/stretchr/testify/assert"
)

type fakeRepoList struct {
	items []*expense.Expense
}

func (f *fakeRepoList) Save(e *expense.Expense) error {
	f.items = append(f.items, e)

	return nil
}

func (f *fakeRepoList) FindAll() ([]*expense.Expense, error) {
	return f.items, nil
}

func TestListExpensesHandlerReturnJsonList(t *testing.T) {
	repo := &fakeRepoList{}
	createdAt := &time.Time{}
	e1, _ := expense.NewExpense(1000, "Food", "Lunch and cafe", createdAt)
	e2, _ := expense.NewExpense(2000, "Transport", "Taxi", createdAt)
	_ = repo.Save(e1)
	_ = repo.Save(e2)

	useCase := app.NewListExpensesUseCase(repo)
	h := handler.NewListExpensesHandler(useCase)

	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp []map[string]any

	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "Food", resp[0]["category"])
	assert.Equal(t, "Transport", resp[1]["category"])
	assert.Equal(t, createdAt.Format(time.RFC3339), resp[0]["createdAt"])
	assert.Equal(t, createdAt.Format(time.RFC3339), resp[1]["createdAt"])
}
