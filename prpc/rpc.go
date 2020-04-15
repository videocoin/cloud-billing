package prpc

import (
	"context"

	pv1 "github.com/videocoin/cloud-api/billing/private/v1"
	v1 "github.com/videocoin/cloud-api/billing/v1"
	"github.com/videocoin/cloud-api/rpc"
	"github.com/videocoin/cloud-billing/datastore"
)

func (s *Server) GetProfileByUserID(ctx context.Context, req *pv1.ProfileRequest) (*v1.ProfileResponse, error) {
	logger := s.logger.WithField("user_id", req.UserID)

	profile := new(v1.ProfileResponse)

	account, err := s.dm.GetAccountByUserID(ctx, req.UserID)
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

func (s *Server) GetCharges(ctx context.Context, req *pv1.ChargesRequest) (*v1.ChargesResponse, error) {
	logger := s.logger.WithField("user_id", req.UserID)

	resp := &v1.ChargesResponse{
		Items: []*v1.ChargeResponse{},
	}

	account, err := s.dm.GetAccountByUserID(ctx, req.UserID)
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
