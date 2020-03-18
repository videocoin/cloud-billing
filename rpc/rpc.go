package rpc

import (
	"context"
	
	"github.com/stripe/stripe-go"
	stripeSess "github.com/stripe/stripe-go/checkout/session"
	v1 "github.com/videocoin/cloud-api/billing/v1"
	"github.com/videocoin/cloud-api/rpc"
)

func (s *Server) MakePayment(ctx context.Context, req *v1.MakePaymentRequest) (*v1.MakePaymentResponse, error) {
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				Name:        stripe.String("Videcoin Payment"),
				Description: stripe.String("Videcoin Payment Description"),
				Amount:      stripe.Int64(req.Amount),
				Currency:    stripe.String(string(stripe.CurrencyUSD)),
				Quantity:    stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String("https://studio.dev.videcoin.network/billing/payments/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("https://studio.dev.videcoin.network/billing/payments/cancel"),
	}

	session, err := stripeSess.New(params)
	if err != nil {
		s.logger.Errorf("failed to stripe session: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	return &v1.MakePaymentResponse{
		SessionId: session.ID,
	}, nil
}

func (s *Server) GetTransactions(ctx context.Context, req *v1.TransactionRequest) (*v1.TransactionListResponse, error) {
	return &v1.TransactionListResponse{
		Items: []*v1.TransactionResponse{},
	}, nil
}
