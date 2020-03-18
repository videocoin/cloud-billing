package rpc

import (
	"context"

	v1 "github.com/videocoin/cloud-api/billing/v1"
)

func (s *Server) MakePayment(ctx context.Context, req *v1.MakePaymentRequest) (*v1.MakePaymentResponse, error) {
	return &v1.MakePaymentResponse{}, nil
}

func (s *Server) GetTransactions(ctx context.Context, req *v1.TransactionRequest) (*v1.TransactionListResponse, error) {
	return &v1.TransactionListResponse{
		Items: []*v1.TransactionResponse{},
	}, nil
}
