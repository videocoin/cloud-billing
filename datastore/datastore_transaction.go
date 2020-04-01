package datastore

import (
	"context"
	"errors"
	"time"

	"github.com/mailru/dbr"
	"github.com/videocoin/cloud-pkg/dbrutil"
	"github.com/videocoin/cloud-pkg/uuid4"
)

var (
	ErrTxNotFound = errors.New("transaction not found")
)

type TransactionDatastore struct {
	conn  *dbr.Connection
	table string
}

func NewTransactionDatastore(conn *dbr.Connection) (*TransactionDatastore, error) {
	return &TransactionDatastore{
		conn:  conn,
		table: "billing_transactions",
	}, nil
}

func (ds *TransactionDatastore) Create(ctx context.Context, transaction *Transaction) error {
	tx, ok := dbrutil.DbTxFromContext(ctx)
	if !ok {
		sess := ds.conn.NewSession(nil)
		tx, err := sess.Begin()
		if err != nil {
			return err
		}

		defer func() {
			err = tx.Commit()
			tx.RollbackUnlessCommitted()
		}()
	}

	if transaction.ID == "" {
		id, err := uuid4.New()
		if err != nil {
			return err
		}

		transaction.ID = id
	}

	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now()
	}

	cols := []string{
		"id", "from", "to", "created_at", "status", "amount",
		"payment_intent_secret", "payment_intent_id", "payment_status",
		"stream_id", "profile_id"}
	_, err := tx.InsertInto(ds.table).Columns(cols...).Record(transaction).Exec()
	if err != nil {
		return err
	}

	return nil
}
