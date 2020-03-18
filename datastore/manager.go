package datastore

import (
	"context"

	"github.com/mailru/dbr"
	"github.com/sirupsen/logrus"
)

type DataManager struct {
	logger *logrus.Entry
	ds     *Datastore
}

func NewDataManager(logger *logrus.Entry, ds *Datastore) (*DataManager, error) {
	return &DataManager{
		logger: logger,
		ds:     ds,
	}, nil
}

func (m *DataManager) NewContext(ctx context.Context) (context.Context, *dbr.Session, *dbr.Tx, error) {
	dbLogger := NewDatastoreLogger(m.logger)
	sess := m.ds.conn.NewSession(dbLogger)
	tx, err := sess.Begin()
	if err != nil {
		return ctx, nil, nil, err
	}

	ctx = NewContextWithDbSession(ctx, sess)
	ctx = NewContextWithDbTx(ctx, tx)

	return ctx, sess, tx, err
}
