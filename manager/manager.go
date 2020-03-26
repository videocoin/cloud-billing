package manager

import (
	"context"

	"github.com/mailru/dbr"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/videocoin/cloud-billing/datastore"
)

type Manager struct {
	logger *logrus.Entry
	ds     *datastore.Datastore
}

func New(opts ...Option) (*Manager, error) {
	ds := &Manager{}
	for _, o := range opts {
		if err := o(ds); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

func (m *Manager) NewContext(ctx context.Context) (context.Context, *dbr.Session, *dbr.Tx, error) {
	dbLogger := datastore.NewDatastoreLogger(m.logger)
	sess := m.ds.NewSession(dbLogger)
	tx, err := sess.Begin()
	if err != nil {
		return ctx, nil, nil, err
	}

	ctx = datastore.NewContextWithDbSession(ctx, sess)
	ctx = datastore.NewContextWithDbTx(ctx, tx)

	return ctx, sess, tx, err
}

func (m *Manager) CreateAccount(ctx context.Context, account *datastore.Account) error {
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

func (m *Manager) GetAccountByUserID(ctx context.Context, userID string) (*datastore.Account, error) {
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

func (m *Manager) CreateTransaction(ctx context.Context, transaction *datastore.Transaction) error {
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

func (m *Manager) UpdateTransactionPaymentIntent(ctx context.Context, transaction *datastore.Transaction, paymentIntent *stripe.PaymentIntent) error {
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

func (m *Manager) GetTransactionByCheckoutSessionID(ctx context.Context, checkoutSessionID string) (*datastore.Transaction, error) {
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
