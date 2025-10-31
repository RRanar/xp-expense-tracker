// Package http keeps all needed logic for HTTP flow processing
package http

import (
	"encoding/json"
	"errors"
	"net/http"

	expense "github.com/RRanar/xp-expense-tracker/internal/domain/expence"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var res ErrorResponse
	switch {
	case errors.Is(err, expense.ErrInvalidAmount):
		w.WriteHeader(http.StatusBadRequest)
		res.Error.Code = "INVALID_AMOUNT"
		res.Error.Message = "Amount must be positive"
	case errors.Is(err, expense.ErrMissingCategory):
		w.WriteHeader(http.StatusBadRequest)
		res.Error.Code = "MISSING_CATEGORY"
		res.Error.Message = "Category is required"
	default:
		w.WriteHeader(http.StatusInternalServerError)
		res.Error.Code = "INTERNAL_ERROR"
		res.Error.Message = "Unexpected error occurred"
	}

	json.NewEncoder(w).Encode(res)
}
