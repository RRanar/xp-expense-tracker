// Package expense defines buisness logic for Expense entity
// defines basic repository with certain buisness actions
// and add basic validation for it
package expense

type Repository interface {
	Save(e *Expense) error
	FindAll() ([]*Expense, error)
}
