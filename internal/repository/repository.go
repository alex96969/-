package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/alex96969/-/ledger/migrations"
	"io/fs"
	_ "modernc.org/sqlite"
	"sort"
)

var ErrNotFound = errors.New("record not found")
var ErrEmailExists = errors.New("email already registered")

type ExpenseRepository struct{ db *sql.DB }

func Open(path string) (*ExpenseRepository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// One connection ensures PRAGMAs also apply to all subsequent queries.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	r := &ExpenseRepository{db: db}
	for _, query := range []string{"PRAGMA foreign_keys = ON", "PRAGMA busy_timeout = 5000", "PRAGMA journal_mode = WAL"} {
		if _, err = db.Exec(query); err != nil {
			db.Close()
			return nil, err
		}
	}
	if err = r.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return r, nil
}

func (r *ExpenseRepository) Close() error { return r.db.Close() }

func (r *ExpenseRepository) migrate() error {
	if _, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY)`); err != nil {
		return err
	}
	names, err := fs.Glob(migrations.Files, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		tx, err := r.db.Begin()
		if err != nil {
			return err
		}
		var exists int
		err = tx.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE name = ?", name).Scan(&exists)
		if err != nil {
			tx.Rollback()
			return err
		}
		if exists != 0 {
			tx.Rollback()
			continue
		}
		script, err := migrations.Files.ReadFile(name)
		if err == nil {
			_, err = tx.Exec(string(script))
		}
		if err == nil {
			_, err = tx.Exec("INSERT INTO schema_migrations(name) VALUES (?)", name)
		}
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
