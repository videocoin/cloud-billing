package datastore

import (
	"time"

	"github.com/mailru/dbr"
	v1 "github.com/videocoin/cloud-api/billing/v1"
)

const (
	BankAccountID = "bank"
)

type Transaction struct {
	ID        string               `db:"id"`
	From      string               `db:"from"`
	To        string               `db:"to"`
	CreatedAt time.Time            `db:"created_at"`
	Amount    int64                `db:"amount"`
	Status    v1.TransactionStatus `db:"status"`

	PaymentIntentSecret dbr.NullString `db:"payment_intent_secret"`
	PaymentIntentID     dbr.NullString `db:"payment_intent_id"`
	PaymentStatus       dbr.NullString `db:"payment_status"`

	StreamID  dbr.NullString `db:"stream_id"`
	ProfileID dbr.NullString `db:"profile_id"`

	CheckedAt time.Time `db:"checked_at"`
	IsLocked  bool      `db:"is_locked"`
}
