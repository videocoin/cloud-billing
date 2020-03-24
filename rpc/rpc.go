package rpc

import (
	"context"
	"fmt"

	protoempty "github.com/gogo/protobuf/types"
	"github.com/stripe/stripe-go"
	checkout "github.com/stripe/stripe-go/checkout/session"
	"github.com/stripe/stripe-go/paymentintent"
	v1 "github.com/videocoin/cloud-api/billing/v1"
	"github.com/videocoin/cloud-api/rpc"
)

func (s *Server) MakePayment(ctx context.Context, req *v1.MakePaymentRequest) (*v1.MakePaymentResponse, error) {
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				Name:        stripe.String("Videcoin Payment"),
				Description: stripe.String("Videcoin Payment Description"),
				Amount:      stripe.Int64(req.Amount),
				Currency:    stripe.String(string(stripe.CurrencyUSD)),
				Quantity:    stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(
			fmt.Sprintf(
				"%s/api/v1/billing/stripe/{CHECKOUT_SESSION_ID}/success",
				s.stripeOpts.BaseCallbackURL,
			)),
		CancelURL: stripe.String(
			fmt.Sprintf(
				"%s/api/v1/billing/stripe/cancel",
				s.stripeOpts.BaseCallbackURL,
			)),
	}

	session, err := checkout.New(params)
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

func (s *Server) SuccessStripeCallback(ctx context.Context, req *v1.StripePaymentRequest) (*protoempty.Empty, error) {
	logger := s.logger.WithField("session_id", req.SessionId)
	logger.Info("stripe payment succeed")

	session, err := checkout.Get(req.SessionId, nil)
	if err != nil {
		logger.Errorf("failed to get checkout: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	if session.PaymentIntent != nil && session.PaymentIntent.ID != "" {
		logger := logger.WithField("payment_intent_id", session.PaymentIntent.ID)

		pi, err := paymentintent.Get(session.PaymentIntent.ID, nil)
		if err != nil {
			logger.Errorf("failed to get payment intent: %s", err)
			return nil, rpc.ErrRpcBadRequest
		}

		session.PaymentIntent = pi

		logger.Infof("payment intent status is %s", pi.Status)

		if pi.Status == stripe.PaymentIntentStatusSucceeded {
		}
	}

	return &protoempty.Empty{}, nil
}

func (s *Server) CancelStripeCallback(ctx context.Context, req *protoempty.Empty) (*protoempty.Empty, error) {
	return &protoempty.Empty{}, nil
}
