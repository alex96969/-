package handlers

import (
	"context"
	"time"

	"github.com/alex96969/-/ledger/internal/models"
)

// Store describes the persistence operations required by the HTTP layer.
// The SQLite implementation is supplied when the application starts.
type Store interface {
	Categories(context.Context) ([]models.Category, error)
	GetAll(context.Context, int64, models.Filter) ([]models.Expense, error)
	GetByID(context.Context, int64, int64) (models.Expense, error)
	Create(context.Context, int64, models.Expense) (int64, error)
	Update(context.Context, int64, int64, models.Expense) error
	Delete(context.Context, int64, int64) error
	CreateUser(context.Context, string, []byte) (int64, error)
	UserByEmail(context.Context, string) (models.User, error)
	CreateSession(context.Context, string, int64, time.Time) error
	SessionUser(context.Context, string, time.Time) (models.User, error)
	DeleteSession(context.Context, string) error
	ByCategory(context.Context, int64, models.Filter) ([]models.Stat, error)
	ByMonth(context.Context, int64, int) ([]models.Stat, error)
}
