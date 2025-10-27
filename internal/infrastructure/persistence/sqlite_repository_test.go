package persistence_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	expense "github.com/RRanar/xp-expense-tracker/internal/domain/expence"
	"github.com/RRanar/xp-expense-tracker/internal/infrastructure/persistence"
	"github.com/stretchr/testify/assert"
)

func TestSQLiteRepositorySaveAndFindAll(t *testing.T) {
	dbFile := "test_expenses.db"
	defer os.Remove(dbFile)
	db, err := sql.Open("sqlite3", dbFile)
	assert.NoError(t, err)
	defer db.Close()
	repo, err := persistence.NewSQLiteRepository(db)
	assert.NoError(t, err)

	createdAt := &time.Time{}
	e1, _ := expense.NewExpense(1000, "Food", "Pizza", createdAt)
	e2, _ := expense.NewExpense(2000, "Transport", "Taxi", createdAt)
	err = repo.Save(e1)
	assert.NoError(t, err)
	err = repo.Save(e2)
	assert.NoError(t, err)

	all, err := repo.FindAll()
	assert.NoError(t, err)
	assert.Len(t, all, 2)
	assert.Equal(t, "Food", all[0].Category)
	assert.Equal(t, "Transport", all[1].Category)
	assert.Equal(t, createdAt, all[0].CreatedAt)
	assert.Equal(t, createdAt, all[1].CreatedAt)
}
