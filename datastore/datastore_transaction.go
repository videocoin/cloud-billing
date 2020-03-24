package datastore

import (
	"context"
	"errors"
	"time"

	"github.com/mailru/dbr"
	"github.com/stripe/stripe-go"
	"github.com/videocoin/cloud-pkg/uuid4"
)

var (
	ErrTxNotFound = errors.New("transaction is not found")
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
	var err error

	sess, ok := DbSessionFromContext(ctx)
	if !ok || sess == nil {
		sess = ds.conn.NewSession(nil)
	}

	tx, ok := DbTxFromContext(ctx)
	if !ok || tx == nil {
		tx, err = sess.Begin()
		if err != nil {
			return err
		}

		defer func() {
			tx.Commit()
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

	cols := []string{"id", "account_id", "created_at", "type", "checkout_session_id", "payment_intent_id", "payment_status", "amount"}
	_, err = tx.InsertInto(ds.table).Columns(cols...).Record(transaction).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (ds *TransactionDatastore) UpdatePaymentIntent(ctx context.Context, transaction *Transaction, paymentIntent *stripe.PaymentIntent) error {
	var err error

	sess, ok := DbSessionFromContext(ctx)
	if !ok || sess == nil {
		sess = ds.conn.NewSession(nil)
	}

	tx, ok := DbTxFromContext(ctx)
	if !ok || tx == nil {
		tx, err = sess.Begin()
		if err != nil {
			return err
		}

		defer func() {
			tx.Commit()
			tx.RollbackUnlessCommitted()
		}()
	}

	transaction.PaymentIntentID = paymentIntent.ID
	transaction.PaymentStatus = paymentIntent.Status

	_, err = tx.
		Update(ds.table).
		Where("id = ?", transaction.ID).
		Set("payment_intent_id", transaction.PaymentIntentID).
		Set("payment_status", transaction.PaymentStatus).
		Exec()

	return err
}

func (ds *TransactionDatastore) GetByCheckoutSessionID(ctx context.Context, checkoutSessionID string) (*Transaction, error) {
	var err error

	sess, ok := DbSessionFromContext(ctx)
	if !ok || sess == nil {
		sess = ds.conn.NewSession(nil)
	}

	tx, ok := DbTxFromContext(ctx)
	if !ok || tx == nil {
		tx, err = sess.Begin()
		if err != nil {
			return nil, err
		}

		defer func() {
			tx.Commit()
			tx.RollbackUnlessCommitted()
		}()
	}

	transaction := new(Transaction)
	_, err = tx.Select("*").From(ds.table).Where("checkout_session_id = ?", checkoutSessionID).Load(transaction)
	if err != nil {
		if err == dbr.ErrNotFound {
			return nil, ErrTxNotFound
		}
		return nil, err
	}

	return transaction, nil
}
