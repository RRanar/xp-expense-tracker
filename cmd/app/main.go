package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	app "github.com/RRanar/xp-expense-tracker/internal/application/usecase"
	handler "github.com/RRanar/xp-expense-tracker/internal/infrastructure/http"
	persistence "github.com/RRanar/xp-expense-tracker/internal/infrastructure/persistence"
)

func main() {
	db, err := sql.Open("sqlite3", "expenses.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo, err := persistence.NewSQLiteRepository(db)
	if err != nil {
		log.Fatal(err)
	}

	createUC := app.NewCreateExpenseUseCase(repo)
	listUC := app.NewListExpensesUseCase(repo)
	createHandler := handler.NewExpenseHandler(createUC)
	listHandler := handler.NewListExpensesHandler(listUC)

	http.Handle("/expenses", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createHandler.ServeHTTP(w, r)
		case http.MethodGet:
			listHandler.ServeHTTP(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
