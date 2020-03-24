package datastore

import (
	"context"
	"errors"
	"time"

	"github.com/mailru/dbr"
	"github.com/videocoin/cloud-pkg/uuid4"
)

var (
	ErrAccountNotFound = errors.New("account is not found")
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

	cols := []string{"id", "user_id", "created_at", "updated_at", "balance"}
	_, err = tx.InsertInto(ds.table).Columns(cols...).Record(account).Exec()
	if err != nil {
		return err
	}

	return nil
}
