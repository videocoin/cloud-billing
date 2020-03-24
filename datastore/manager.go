package datastore

import (
	"context"

	"github.com/mailru/dbr"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
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

func (m *DataManager) CreateAccount(ctx context.Context, account *Account) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return failedTo("create account", err)
	}
	defer tx.RollbackUnlessCommitted()

	err = m.ds.Accounts.Create(ctx, account)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *DataManager) GetAccountByUserID(ctx context.Context, userID string) (*Account, error) {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return nil, failedTo("get account by user id", err)
	}
	defer tx.RollbackUnlessCommitted()

	account, err := m.ds.Accounts.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (m *DataManager) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return failedTo("create transaction", err)
	}
	defer tx.RollbackUnlessCommitted()

	err = m.ds.Transactions.Create(ctx, transaction)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *DataManager) UpdateTransactionPaymentIntent(ctx context.Context, transaction *Transaction, paymentIntent *stripe.PaymentIntent) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return failedTo("update transaction payment intent", err)
	}
	defer tx.RollbackUnlessCommitted()

	err = m.ds.Transactions.UpdatePaymentIntent(ctx, transaction, paymentIntent)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *DataManager) GetTransactionByCheckoutSessionID(ctx context.Context, checkoutSessionID string) (*Transaction, error) {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return nil, failedTo("get transaction by checkout session id", err)
	}
	defer tx.RollbackUnlessCommitted()

	transaction, err := m.ds.Transactions.GetByCheckoutSessionID(ctx, checkoutSessionID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
