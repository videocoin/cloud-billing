package datastore

import (
	_ "github.com/go-sql-driver/mysql" //nolint
	"github.com/mailru/dbr"
)

type Datastore struct {
	conn *dbr.Connection
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

	return ds, nil
}
