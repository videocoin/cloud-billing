package datastore

import (
	"time"

	"github.com/stripe/stripe-go"
	v1 "github.com/videocoin/cloud-api/billing/v1"
)

type Transaction struct {
	ID                string                     `db:"id"`
	AccountID         string                     `db:"account_id"`
	CreatedAt         time.Time                  `db:"created_at"`
	Type              v1.TransactionType         `db:"type"`
	CheckoutSessionID string                     `db:"checkout_session_id"`
	PaymentIntentID   string                     `db:"payment_intent_id"`
	PaymentStatus     stripe.PaymentIntentStatus `db:"payment_status"`
	Amount            int64                      `db:"amount"`
}
