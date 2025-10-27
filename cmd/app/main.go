package main

import (
	"log"
	"net/http"

	app "github.com/RRanar/xp-expense-tracker/internal/application/usecase"
	"github.com/RRanar/xp-expense-tracker/internal/domain/expence"
	handler "github.com/RRanar/xp-expense-tracker/internal/infrastructure/http"
)

type inMemoryRepo struct {
	saved []*expense.Expense
}

func (r *inMemoryRepo) Save(e *expense.Expense) error {
	r.saved = append(r.saved, e)

	return nil
}

func (r *inMemoryRepo) FindAll() ([]*expense.Expense, error) {
	return r.saved, nil
}

func main() {
	repo := &inMemoryRepo{}
	h := handler.NewExpenseHandler(app.NewCreateExpenseUseCase(repo))
	listHandler := handler.NewListExpensesHandler(app.NewListExpensesUseCase(repo))

	http.Handle("/expenses", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.ServeHTTP(w, r)
		case http.MethodGet:
			listHandler.ServeHTTP(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
