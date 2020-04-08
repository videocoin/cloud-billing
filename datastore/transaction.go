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
	Amount    float64              `db:"amount"`
	Status    v1.TransactionStatus `db:"status"`

	CheckedAt time.Time `db:"checked_at"`
	IsLocked  bool      `db:"is_locked"`

	PaymentIntentSecret dbr.NullString `db:"payment_intent_secret"`
	PaymentIntentID     dbr.NullString `db:"payment_intent_id"`
	PaymentStatus       dbr.NullString `db:"payment_status"`

	StreamID              dbr.NullString `db:"stream_id"`
	StreamName            dbr.NullString `db:"stream_name"`
	StreamContractAddress dbr.NullString `db:"stream_contract_address"`
	StreamIsLive          bool           `db:"stream_is_live"`

	ProfileID   dbr.NullString  `db:"profile_id"`
	ProfileName dbr.NullString  `db:"profile_name"`
	ProfileCost dbr.NullFloat64 `db:"profile_cost"`

	TaskID   dbr.NullString `db:"task_id"`
	ChunkNum dbr.NullInt64  `db:"chunk_num"`
	Duration dbr.NullInt64  `db:"duration"`

	Price dbr.NullFloat64 `db:"price"`
}
