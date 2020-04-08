package rpc

import (
	"context"

	prototypes "github.com/gogo/protobuf/types"
	"github.com/mailru/dbr"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/paymentintent"
	v1 "github.com/videocoin/cloud-api/billing/v1"
	"github.com/videocoin/cloud-api/rpc"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-billing/datastore"
)

func (s *Server) GetProfile(ctx context.Context, req *prototypes.Empty) (*v1.ProfileResponse, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithField("user_id", userID)

	profile := new(v1.ProfileResponse)

	account, err := s.dm.GetAccountByUserID(ctx, userID)
	if err != nil {
		if err == datastore.ErrAccountNotFound {
			return profile, nil
		}
		logger.Errorf("failed to get account by user id: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	balance, err := s.dm.GetBalance(ctx, account)
	if err != nil {
		logger.Errorf("failed to get balance: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	profile.Balance = balance

	return profile, nil
}

func (s *Server) MakePayment(ctx context.Context, req *v1.MakePaymentRequest) (*v1.MakePaymentResponse, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	amount := req.Amount * 100

	logger := s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"amount":  amount,
	})

	account, err := s.dm.GetAccountByUserID(ctx, userID)
	if err != nil {
		if err == datastore.ErrAccountNotFound {
			user, err := s.users.GetById(ctx, &usersv1.UserRequest{Id: userID})
			if err != nil {
				logger.Errorf("failed to get user: %s", err)
				return nil, rpc.ErrRpcInternal
			}

			account = &datastore.Account{UserID: user.ID, Email: user.Email}
			createErr := s.dm.CreateAccount(ctx, account)
			if createErr != nil {
				logger.Errorf("failed to create account: %s", createErr)
				return nil, rpc.ErrRpcInternal
			}
		} else {
			logger.Errorf("failed to get account by user id: %s", err)
			return nil, rpc.ErrRpcInternal
		}
	}

	if account.CustomerID.String == "" {
		cus, err := customer.New(&stripe.CustomerParams{
			Email: stripe.String(account.Email),
		})
		if err != nil {
			logger.Errorf("failed to create stripe customer: %s", err)
			return nil, rpc.ErrRpcInternal
		}

		err = s.dm.UpdateAccountCustomer(ctx, account, cus.ID)
		if err != nil {
			logger.Errorf("failed to update account customer: %s", err)
			return nil, rpc.ErrRpcInternal
		}
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		Customer: stripe.String(account.CustomerID.String),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		s.logger.Errorf("failed to new payment intent: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	transaction := &datastore.Transaction{
		From:                datastore.BankAccountID,
		To:                  account.ID,
		Amount:              float64(amount),
		Status:              v1.TransactionStatusProcesing,
		PaymentIntentID:     dbr.NewNullString(pi.ID),
		PaymentIntentSecret: dbr.NewNullString(pi.ClientSecret),
		PaymentStatus:       dbr.NewNullString(pi.Status),
	}

	err = s.dm.CreateTransaction(ctx, transaction)
	if err != nil {
		s.logger.Errorf("failed to create transaction: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	return &v1.MakePaymentResponse{
		ClientSecret: pi.ClientSecret,
	}, nil
}

func (s *Server) GetCharges(ctx context.Context, req *prototypes.Empty) (*v1.ChargesResponse, error) {
	resp := &v1.ChargesResponse{
		Items: []*v1.ChargeResponse{},
	}

	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithField("user_id", userID)

	account, err := s.dm.GetAccountByUserID(ctx, userID)
	if err != nil {
		if err == datastore.ErrAccountNotFound {
			return resp, nil
		}
		logger.Errorf("failed to get account by user id: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	charges, err := s.dm.GetCharges(ctx, account)
	if err != nil {
		logger.Errorf("failed to get charges: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	resp.Items = charges

	return resp, nil
}

func (s *Server) GetTransactions(ctx context.Context, req *prototypes.Empty) (*v1.TransactionsResponse, error) {
	resp := &v1.TransactionsResponse{
		Items: []*v1.TransactionResponse{},
	}

	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithField("user_id", userID)

	account, err := s.dm.GetAccountByUserID(ctx, userID)
	if err != nil {
		if err == datastore.ErrAccountNotFound {
			return resp, nil
		}
		logger.Errorf("failed to get account by user id: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	transactions, err := s.dm.GetTransactions(ctx, account)
	if err != nil {
		logger.Errorf("failed to get transactions: %s", err)
		return nil, rpc.ErrRpcInternal
	}

	resp.Items = transactions

	return resp, nil
}
