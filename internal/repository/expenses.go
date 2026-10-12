package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/alex96969/-/ledger/internal/models"
)

func (r *ExpenseRepository) Categories(ctx context.Context) ([]models.Category, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM categories ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Category{}
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func conditions(userID int64, f models.Filter) (string, []any) {
	q := " WHERE e.user_id = ?"
	args := []any{userID}
	if f.CategoryID != 0 {
		q += " AND e.category_id = ?"
		args = append(args, f.CategoryID)
	}
	if f.From != "" {
		q += " AND e.date >= ?"
		args = append(args, f.From)
	}
	if f.To != "" {
		q += " AND e.date <= ?"
		args = append(args, f.To)
	}
	return q, args
}

const expenseSelect = `SELECT e.id, e.user_id, e.amount, e.description, e.date, e.category_id, c.name FROM expenses e JOIN categories c ON c.id = e.category_id`

func (r *ExpenseRepository) GetAll(ctx context.Context, userID int64, f models.Filter) ([]models.Expense, error) {
	where, args := conditions(userID, f)
	rows, err := r.db.QueryContext(ctx, expenseSelect+where+orderBy(f.Sort), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Expense{}
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.UserID, &e.Amount, &e.Description, &e.Date, &e.CategoryID, &e.Category); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *ExpenseRepository) GetByID(ctx context.Context, userID, id int64) (models.Expense, error) {
	var e models.Expense
	err := r.db.QueryRowContext(ctx, expenseSelect+" WHERE e.user_id = ? AND e.id = ?", userID, id).Scan(&e.ID, &e.UserID, &e.Amount, &e.Description, &e.Date, &e.CategoryID, &e.Category)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNotFound
	}
	return e, err
}

func (r *ExpenseRepository) Create(ctx context.Context, userID int64, e models.Expense) (int64, error) {
	res, err := r.db.ExecContext(ctx, "INSERT INTO expenses(user_id,amount,description,date,category_id) VALUES (?,?,?,?,?)", userID, e.Amount, e.Description, e.Date, e.CategoryID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func changed(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *ExpenseRepository) Update(ctx context.Context, userID, id int64, e models.Expense) error {
	return changed(r.db.ExecContext(ctx, "UPDATE expenses SET amount=?,description=?,date=?,category_id=? WHERE user_id=? AND id=?", e.Amount, e.Description, e.Date, e.CategoryID, userID, id))
}
func (r *ExpenseRepository) Delete(ctx context.Context, userID, id int64) error {
	return changed(r.db.ExecContext(ctx, "DELETE FROM expenses WHERE user_id=? AND id=?", userID, id))
}

// Only fixed SQL fragments are used; user input is never interpolated into ORDER BY.
func orderBy(sort string) string {
	switch sort {
	case "oldest":
		return " ORDER BY e.date ASC, e.id ASC"
	case "largest":
		return " ORDER BY e.amount DESC, e.id DESC"
	case "smallest":
		return " ORDER BY e.amount ASC, e.id DESC"
	default:
		return " ORDER BY e.date DESC, e.id DESC"
	}
}
