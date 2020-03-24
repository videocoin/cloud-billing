package datastore

import (
	_ "github.com/go-sql-driver/mysql" //nolint
	"github.com/mailru/dbr"
)

type Datastore struct {
	conn *dbr.Connection

	Accounts     *AccountDatastore
	Transactions *TransactionDatastore
}

func NewDatastore(uri string) (*Datastore, error) {
	ds := new(Datastore)

	conn, err := dbr.Open("mysql", uri, nil)
	if err != nil {
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	ds.conn = conn

	accountDs, err := NewAccountDatastore(conn)
	if err != nil {
		return nil, err
	}

	ds.Accounts = accountDs

	txDs, err := NewTransactionDatastore(conn)
	if err != nil {
		return nil, err
	}

	ds.Transactions = txDs

	return ds, nil
}
