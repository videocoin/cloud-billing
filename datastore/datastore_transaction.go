package datastore

import (
	"context"
	"errors"
	"time"

	"github.com/mailru/dbr"
	"github.com/stripe/stripe-go"
	v1 "github.com/videocoin/cloud-api/billing/v1"
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

func (ds *TransactionDatastore) markStatusAs(ctx context.Context, transaction *Transaction, status v1.TransactionStatus) error {
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

	transaction.Status = status

	isUpdatePaymentStatus := true

	switch status {
	case v1.TransactionStatusSuccess:
		transaction.PaymentStatus = dbr.NewNullString(stripe.PaymentIntentStatusSucceeded)
	case v1.TransactionStatusCanceled:
		transaction.PaymentStatus = dbr.NewNullString(stripe.PaymentIntentStatusCanceled)
	default:
		isUpdatePaymentStatus = false
	}

	b := tx.
		Update(ds.table).
		Where("id = ?", transaction.ID).
		Set("status", transaction.Status)

	if isUpdatePaymentStatus {
		b = b.Set("payment_status", transaction.PaymentStatus)
	}

	_, err := b.Exec()
	if err != nil {
		return err
	}

	return nil
}

func (ds *TransactionDatastore) markPaymentStatusAs(ctx context.Context, transaction *Transaction, status stripe.PaymentIntentStatus) error {
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

	transaction.PaymentStatus = dbr.NewNullString(status)

	_, err := tx.
		Update(ds.table).
		Where("id = ?", transaction.ID).
		Set("payment_status", transaction.PaymentStatus).
		Exec()
	if err != nil {
		return err
	}

	return nil
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

	if transaction.CheckedAt.IsZero() {
		transaction.CheckedAt = time.Now()
	}

	cols := []string{
		"id", "from", "to", "created_at", "status", "amount",
		"payment_intent_secret", "payment_intent_id", "payment_status",
		"stream_id", "profile_id", "checked_at", "is_locked"}
	_, err := tx.InsertInto(ds.table).Columns(cols...).Record(transaction).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (ds *TransactionDatastore) GetToCheckPayment(ctx context.Context) (*Transaction, error) {
	tx, ok := dbrutil.DbTxFromContext(ctx)
	if !ok {
		sess := ds.conn.NewSession(nil)
		tx, err := sess.Begin()
		if err != nil {
			return nil, err
		}

		defer func() {
			err = tx.Commit()
			tx.RollbackUnlessCommitted()
		}()
	}

	transaction := new(Transaction)

	err := tx.
		Select("*").
		ForUpdate().
		From(ds.table).
		Where("is_locked = ? AND status = ? AND payment_intent_id IS NOT NULL", false, v1.TransactionStatusProcesing).
		OrderBy("checked_at").
		Limit(1).
		LoadStruct(transaction)
	if err != nil {
		if err == dbr.ErrNotFound {
			return nil, ErrTxNotFound
		}
		return nil, err
	}

	_, err = tx.Update(ds.table).Where("id = ?", transaction.ID).Set("is_locked", true).Exec()
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (ds *TransactionDatastore) GetByPaymentID(ctx context.Context, id string) (*Transaction, error) {
	tx, ok := dbrutil.DbTxFromContext(ctx)
	if !ok {
		sess := ds.conn.NewSession(nil)
		tx, err := sess.Begin()
		if err != nil {
			return nil, err
		}

		defer func() {
			err = tx.Commit()
			tx.RollbackUnlessCommitted()
		}()
	}

	transaction := new(Transaction)
	err := tx.
		Select("*").
		From(ds.table).
		Where("payment_intent_id = ?", id).
		Limit(1).
		LoadStruct(transaction)
	if err != nil {
		if err == dbr.ErrNotFound {
			return nil, ErrTxNotFound
		}
		return nil, err
	}

	return transaction, nil
}

func (ds *TransactionDatastore) UnlockToCheckPayment(ctx context.Context, transaction *Transaction) error {
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

	transaction.IsLocked = false
	transaction.CheckedAt = time.Now()

	_, err := tx.
		Update(ds.table).
		Where("id = ?", transaction.ID).
		Set("is_locked", transaction.IsLocked).
		Set("checked_at", transaction.CheckedAt).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func (ds *TransactionDatastore) UnlockAll(ctx context.Context) error {
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

	_, err := tx.
		Update(ds.table).
		Set("is_locked", false).
		Where("is_locked = ? AND checked_at <= ?", true, time.Now().Add(time.Minute*-1)).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func (ds *TransactionDatastore) MarkAsSucceded(ctx context.Context, transaction *Transaction) error {
	return ds.markStatusAs(ctx, transaction, v1.TransactionStatusSuccess)
}

func (ds *TransactionDatastore) MarkAsCanceled(ctx context.Context, transaction *Transaction) error {
	return ds.markStatusAs(ctx, transaction, v1.TransactionStatusCanceled)
}

func (ds *TransactionDatastore) MarkAsFailed(ctx context.Context, transaction *Transaction) error {
	return ds.markStatusAs(ctx, transaction, v1.TransactionStatusFailed)
}

func (ds *TransactionDatastore) MarkPaymentStatusAs(ctx context.Context, transaction *Transaction, status stripe.PaymentIntentStatus) error {
	return ds.markPaymentStatusAs(ctx, transaction, status)
}

func (ds *TransactionDatastore) CalcBalance(ctx context.Context, account *Account) (float64, error) {
	tx, ok := dbrutil.DbTxFromContext(ctx)
	if !ok {
		sess := ds.conn.NewSession(nil)
		tx, err := sess.Begin()
		if err != nil {
			return 0, err
		}

		defer func() {
			err = tx.Commit()
			tx.RollbackUnlessCommitted()
		}()
	}

	balance := float64(0)
	err := tx.
		Select("SUM(amount)/100").
		From(ds.table).
		Where("`to` = ? AND status = ?", account.ID, v1.TransactionStatusSuccess).
		LoadStruct(&balance)
	if err != nil {
		return 0, err
	}

	return balance, nil
}
