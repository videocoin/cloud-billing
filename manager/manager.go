package manager

import (
	"context"
	"time"

	"github.com/mailru/dbr"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
	"github.com/videocoin/cloud-billing/datastore"
	"github.com/videocoin/cloud-pkg/dbrutil"
)

type Manager struct {
	logger    *logrus.Entry
	ds        *datastore.Datastore
	cpsticker *time.Ticker
	uatticker *time.Ticker
}

func New(opts ...Option) (*Manager, error) {
	ds := &Manager{
		cpsticker: time.NewTicker(time.Second * 5),
		uatticker: time.NewTicker(time.Second * 30),
	}
	for _, o := range opts {
		if err := o(ds); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

func (m *Manager) Start() {
	go m.checkPaymentStatus()
	go m.unlockAllTransactions()

}

func (m *Manager) Stop() {
	m.cpsticker.Stop()
	m.uatticker.Stop()
}

func (m *Manager) checkPaymentStatus() {
	for range m.cpsticker.C {
		m.logger.Info("checking payment intent status")

		ctx := context.Background()

		m.logger.Info("getting transcation to check payment")
		transaction, err := m.GetTransactionToCheckPayment(ctx)
		if err != nil {
			if err == datastore.ErrTxNotFound {
				continue
			}
			m.logger.Errorf("failed to get transaction to check payment: %s", err)
			continue
		}

		logger := m.logger.WithField("tx_id", transaction.ID)

		pi, err := paymentintent.Get(transaction.PaymentIntentID.String, nil)
		if err != nil {
			m.logger.Errorf("failed to get payment intent: %s", err)
			continue
		}

		logger = m.logger.WithField("pi_id", transaction.ID)

		switch pi.Status {
		case stripe.PaymentIntentStatusSucceeded:
			logger.Info("marking transaction as succeeded")
			err = m.MarkTransactionAsSucceded(ctx, transaction)
			if err != nil {
				m.logger.Errorf("failed to mark transaction as succeded: %s", err)
				continue
			}
		case stripe.PaymentIntentStatusCanceled:
			logger.Info("marking transaction as canceled")
			err = m.MarkTransactionAsCanceled(ctx, transaction)
			if err != nil {
				m.logger.Errorf("failed to mark transaction as canceled: %s", err)
				continue
			}
		default:
			logger.Infof("payment intent status is %s", pi.Status)
			if transaction.PaymentStatus.String != string(pi.Status) {
				err = m.MarkTransactionPaymentStatusAs(ctx, transaction, pi.Status)
				if err != nil {
					m.logger.Errorf("failed to mark transaction payment stauts as %s: %s", pi.Status, err)
					continue
				}
			}
		}

		m.logger.Info("unlocking transcation to check payment")
		err = m.UnlockTransactionToCheckPayment(ctx, transaction)
		if err != nil {
			m.logger.Errorf("failed to unlock transaction to check payment: %s", err)
			continue
		}
	}
}

func (m *Manager) unlockAllTransactions() {
	for range m.uatticker.C {
		m.logger.Info("unlocking all locked transactions")
		err := m.UnlockAllTransactions(context.Background())
		if err != nil {
			m.logger.Errorf("failed to unlock all transactions: %s", err)
			continue
		}
	}
}

func (m *Manager) NewContext(ctx context.Context) (context.Context, *dbr.Session, *dbr.Tx, error) {
	dbLogger := dbrutil.NewDatastoreLogger(m.logger)
	sess := m.ds.NewSession(dbLogger)
	tx, err := sess.Begin()
	if err != nil {
		return ctx, nil, nil, err
	}

	ctx = dbrutil.NewContextWithDbSession(ctx, sess)
	ctx = dbrutil.NewContextWithDbTx(ctx, tx)

	return ctx, sess, tx, err
}

func (m *Manager) CreateAccount(ctx context.Context, account *datastore.Account) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return err
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
		return nil, err
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

func (m *Manager) UpdateAccountCustomer(ctx context.Context, account *datastore.Account, customerID string) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = m.ds.Accounts.UpdateCustomer(ctx, account, customerID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Manager) CreateTransaction(ctx context.Context, transaction *datastore.Transaction) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = m.ds.Transactions.Create(ctx, transaction)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Manager) GetTransactionToCheckPayment(ctx context.Context) (*datastore.Transaction, error) {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUnlessCommitted()

	transaction, err := m.ds.Transactions.GetToCheckPayment(ctx)
	if err != nil {
		return nil, err
	}

	return transaction, tx.Commit()
}

func (m *Manager) UnlockTransactionToCheckPayment(ctx context.Context, transaction *datastore.Transaction) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = m.ds.Transactions.UnlockToCheckPayment(ctx, transaction)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Manager) UnlockAllTransactions(ctx context.Context) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = m.ds.Transactions.UnlockAll(ctx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Manager) MarkTransactionAsSucceded(ctx context.Context, transaction *datastore.Transaction) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = m.ds.Transactions.MarkAsSucceded(ctx, transaction)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Manager) MarkTransactionAsCanceled(ctx context.Context, transaction *datastore.Transaction) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = m.ds.Transactions.MarkAsCanceled(ctx, transaction)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Manager) MarkTransactionAsFailed(ctx context.Context, transaction *datastore.Transaction) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = m.ds.Transactions.MarkAsFailed(ctx, transaction)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Manager) MarkTransactionPaymentStatusAs(ctx context.Context, transaction *datastore.Transaction, status stripe.PaymentIntentStatus) error {
	ctx, _, tx, err := m.NewContext(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = m.ds.Transactions.MarkPaymentStatusAs(ctx, transaction, status)
	if err != nil {
		return err
	}

	return tx.Commit()
}
