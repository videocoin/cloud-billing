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
	ErrAccountNotFound = errors.New("account not found")
)

type AccountDatastore struct {
	conn  *dbr.Connection
	table string
}

func NewAccountDatastore(conn *dbr.Connection) (*AccountDatastore, error) {
	return &AccountDatastore{
		conn:  conn,
		table: "billing_accounts",
	}, nil
}

func (ds *AccountDatastore) Create(ctx context.Context, account *Account) error {
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

	if account.ID == "" {
		id, err := uuid4.New()
		if err != nil {
			return err
		}

		account.ID = id
	}

	if account.CreatedAt.IsZero() {
		account.CreatedAt = time.Now()
	}

	cols := []string{"id", "user_id", "email", "created_at", "updated_at", "balance", "customer_id"}
	_, err := tx.InsertInto(ds.table).Columns(cols...).Record(account).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (ds *AccountDatastore) GetByUserID(ctx context.Context, userID string) (*Account, error) {
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

	account := new(Account)
	err := tx.Select("*").From(ds.table).Where("user_id = ?", userID).LoadStruct(account)
	if err != nil {
		if err == dbr.ErrNotFound {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	return account, nil
}

func (ds *AccountDatastore) GetByID(ctx context.Context, id string) (*Account, error) {
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

	account := new(Account)
	err := tx.Select("*").From(ds.table).Where("id = ?", id).LoadStruct(account)
	if err != nil {
		if err == dbr.ErrNotFound {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	return account, nil
}

func (ds *AccountDatastore) UpdateCustomer(ctx context.Context, account *Account, customerID string) error {
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

	account.CustomerID = dbr.NewNullString(customerID)

	_, err := tx.
		Update(ds.table).
		Where("id = ?", account.ID).
		Set("customer_id", account.CustomerID).
		Exec()

	return err
}
