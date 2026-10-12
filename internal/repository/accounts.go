package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/alex96969/-/ledger/internal/models"
	"modernc.org/sqlite"
	"strings"
	"time"
)

func (r *ExpenseRepository) CreateUser(ctx context.Context, email string, hash []byte) (int64, error) {
	res, err := r.db.ExecContext(ctx, "INSERT INTO users(email,password_hash) VALUES (?,?)", strings.ToLower(email), hash)
	if err != nil {
		var se *sqlite.Error
		if errors.As(err, &se) && se.Code() == 2067 {
			return 0, ErrEmailExists
		}
		return 0, err
	}
	return res.LastInsertId()
}
func (r *ExpenseRepository) UserByEmail(ctx context.Context, email string) (models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(ctx, "SELECT id,email,password_hash FROM users WHERE email=?", email).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNotFound
	}
	return u, err
}
func (r *ExpenseRepository) CreateSession(ctx context.Context, hash string, userID int64, expires time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at <= ?", time.Now().Unix()); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO sessions(session_id,user_id,expires_at) VALUES (?,?,?)", hash, userID, expires.Unix()); err != nil {
		return err
	}
	return tx.Commit()
}
func (r *ExpenseRepository) SessionUser(ctx context.Context, hash string, now time.Time) (models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(ctx, `SELECT u.id,u.email FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.session_id=? AND s.expires_at>?`, hash, now.Unix()).Scan(&u.ID, &u.Email)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNotFound
	}
	return u, err
}
func (r *ExpenseRepository) DeleteSession(ctx context.Context, hash string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE session_id=?", hash)
	return err
}
