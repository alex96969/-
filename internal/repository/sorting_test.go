package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alex96969/-/ledger/internal/models"
)

func TestExpenseOrdering(t *testing.T) {
	ctx := context.Background()
	repo, err := Open(filepath.Join(t.TempDir(), "sorting.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	user, err := repo.CreateUser(ctx, "sort@example.test", []byte("hash"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expense := range []models.Expense{{Amount: 300, Description: "first", Date: "2026-01-01", CategoryID: 1}, {Amount: 100, Description: "second", Date: "2026-02-01", CategoryID: 1}, {Amount: 200, Description: "third", Date: "2026-03-01", CategoryID: 2}} {
		if _, err := repo.Create(ctx, user, expense); err != nil {
			t.Fatal(err)
		}
	}
	for sort, want := range map[string]string{"newest": "third", "oldest": "first", "largest": "first", "smallest": "second", "untrusted SQL": "third"} {
		rows, err := repo.GetAll(ctx, user, models.Filter{Sort: sort})
		if err != nil || len(rows) != 3 || rows[0].Description != want {
			t.Fatalf("%s => %+v, %v", sort, rows, err)
		}
	}
	rows, err := repo.GetAll(ctx, user, models.Filter{CategoryID: 1, Sort: "smallest"})
	if err != nil || len(rows) != 2 || rows[0].Amount != 100 {
		t.Fatal(rows, err)
	}
}
