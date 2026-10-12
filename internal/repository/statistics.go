package repository

import (
	"context"
	"fmt"
	"github.com/alex96969/-/ledger/internal/models"
)

func (r *ExpenseRepository) ByCategory(ctx context.Context, userID int64, f models.Filter) ([]models.Stat, error) {
	where, args := conditions(userID, f)
	return r.stats(ctx, `SELECT c.name,SUM(e.amount),COUNT(*) FROM expenses e JOIN categories c ON c.id=e.category_id`+where+` GROUP BY c.id,c.name ORDER BY SUM(e.amount) DESC,c.id`, args...)
}
func (r *ExpenseRepository) ByMonth(ctx context.Context, userID int64, year int) ([]models.Stat, error) {
	rows, err := r.stats(ctx, `SELECT substr(date,1,7),SUM(amount),COUNT(*) FROM expenses WHERE user_id=? AND date>=? AND date<=? GROUP BY substr(date,1,7) ORDER BY 1`, userID, fmt.Sprintf("%04d-01-01", year), fmt.Sprintf("%04d-12-31", year))
	if err != nil {
		return nil, err
	}
	out := make([]models.Stat, 12)
	for i := range out {
		out[i].Label = fmt.Sprintf("%04d-%02d", year, i+1)
		for _, s := range rows {
			if s.Label == out[i].Label {
				out[i] = s
				break
			}
		}
	}
	return out, nil
}
func (r *ExpenseRepository) stats(ctx context.Context, q string, args ...any) ([]models.Stat, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Stat{}
	for rows.Next() {
		var s models.Stat
		if err := rows.Scan(&s.Label, &s.Amount, &s.Count); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
