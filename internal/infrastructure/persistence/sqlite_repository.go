// Package persistence is an infrastructure layer for
// concrete db repository implenations
package persistence

import (
	"database/sql"
	"errors"
	"time"

	expense "github.com/RRanar/xp-expense-tracker/internal/domain/expence"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) (*SQLiteRepository, error) {
	repo := &SQLiteRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *SQLiteRepository) migrate() error {
	schema := `
	  CREATE TABLE expenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			amount INTEGER NOT NULL,
			category TEXT NOT NULL,
			description TEXT,
			created_at DATETIME NOT NULL
		)
	`
	_, err := s.db.Exec(schema)

	return err
}

func (s *SQLiteRepository) Save(e *expense.Expense) error {
	if e == nil {
		return errors.New("expense is null")
	}

	_, err := s.db.Exec(
		"INSERT INTO expenses (amount, category, description, created_at) VALUES (?, ?, ?, ?)",
		e.Amount,
		e.Category,
		e.Description,
		e.CreatedAt.Get(),
	)

	return err
}

func (s *SQLiteRepository) FindAll() ([]*expense.Expense, error) {
	rows, err := s.db.Query("SELECT amount, category, description, created_at FROM expenses ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*expense.Expense
	for rows.Next() {
		var (
			amount      int64
			category    string
			description string
			createdAt   string
		)
		if err := rows.Scan(&amount, &category, &description, &createdAt); err != nil {
			return nil, err
		}

		createdTime, _ := time.Parse(time.RFC3339, createdAt)
		e, err := expense.NewExpense(amount, category, description, &createdTime)
		if err != nil {
			return nil, err
		}

		items = append(items, e)
	}

	return items, nil
}
