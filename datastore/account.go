package datastore

import (
	"time"

	"github.com/mailru/dbr"
)

type Account struct {
	ID         string         `db:"id"`
	UserID     string         `db:"user_id"`
	Email      string         `db:"email"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  *time.Time     `db:"updated_at"`
	Balance    int64          `db:"balance"`
	CustomerID dbr.NullString `db:"customer_id"`
}
