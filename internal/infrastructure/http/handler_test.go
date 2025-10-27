package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	app "github.com/RRanar/xp-expense-tracker/internal/application/usecase"
	"github.com/RRanar/xp-expense-tracker/internal/domain/expence"
	handler "github.com/RRanar/xp-expense-tracker/internal/infrastructure/http"
	"github.com/stretchr/testify/assert"
)

type fakeRepo struct {
	saved []*expense.Expense
}

func (f *fakeRepo) Save(e *expense.Expense) error {
	f.saved = append(f.saved, e)

	return nil
}

func (f *fakeRepo) FindAll() ([]*expense.Expense, error) {
	return f.saved, nil
}

func TestCreateExpenseHandlerSuccess(t *testing.T) {
	repo := &fakeRepo{}
	useCase := app.NewCreateExpenseUseCase(repo)
	h := handler.NewExpenseHandler(useCase)

	body := map[string]any{
		"amount":      2500,
		"category":    "Food",
		"description": "Lunch and cafe",
	}

	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Food", resp["category"])
}
