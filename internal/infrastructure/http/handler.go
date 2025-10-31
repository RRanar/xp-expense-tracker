package http

import (
	"encoding/json"
	"net/http"

	app "github.com/RRanar/xp-expense-tracker/internal/application/usecase"
)

type ExpenseHandler struct {
	useCase *app.CreateExpenseUseCase
}

func NewExpenseHandler(uc *app.CreateExpenseUseCase) *ExpenseHandler {
	return &ExpenseHandler{useCase: uc}
}

func (h *ExpenseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input app.CreateExpenseInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
	}

	output, err := h.useCase.Execute(input)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(output)
}
