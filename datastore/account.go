package datastore

import (
	"time"
)

type Account struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	Balance   int64      `db:"balance"`
}
