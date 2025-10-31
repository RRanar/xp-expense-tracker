package http

import (
	"encoding/json"
	"net/http"

	app "github.com/RRanar/xp-expense-tracker/internal/application/usecase"
)

type ListExpensesHandler struct {
	useCase *app.ListExpensesUseCase
}

func NewListExpensesHandler(uc *app.ListExpensesUseCase) *ListExpensesHandler {
	return &ListExpensesHandler{useCase: uc}
}

func (h *ListExpensesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	out, err := h.useCase.Execute()
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(out)
}
