package rpc

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"
	accountsv1 "github.com/videocoin/cloud-api/accounts/v1"
	v1 "github.com/videocoin/cloud-api/billing/v1"
	"github.com/videocoin/cloud-api/rpc"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-billing/manager"
	"github.com/videocoin/cloud-pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	logger          *logrus.Entry
	addr            string
	authTokenSecret string
	grpc            *grpc.Server
	listen          net.Listener
	validator       *requestValidator
	dm              *manager.Manager
	accounts        accountsv1.AccountServiceClient
	users           usersv1.UserServiceClient
}

func NewServer(opts ...Option) (*Server, error) {
	s := &Server{
		grpc: grpc.NewServer(),
	}
	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}

	v1.RegisterBillingServiceServer(s.grpc, s)
	reflection.Register(s.grpc)

	return s, nil
}

func (s *Server) Start() error {
	s.logger.Infof("starting rpc server on %s", s.addr)
	return s.grpc.Serve(s.listen)
}

func (s *Server) authenticate(ctx context.Context) (string, error) {
	ctx = auth.NewContextWithSecretKey(ctx, s.authTokenSecret)
	ctx, jwtToken, err := auth.AuthFromContext(ctx)
	if err != nil {
		s.logger.Warningf("failed to auth from context: %s", err)
		return "", rpc.ErrRpcUnauthenticated
	}

	tokenType, ok := auth.TypeFromContext(ctx)
	if ok {
		if usersv1.TokenType(tokenType) == usersv1.TokenTypeAPI {
			_, err := s.users.GetApiToken(context.Background(), &usersv1.ApiTokenRequest{Token: jwtToken})
			if err != nil {
				s.logger.Errorf("failed to get api token: %s", err)
				return "", rpc.ErrRpcUnauthenticated
			}
		}
	}

	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		s.logger.Warningf("failed to get user id from context: %s", err)
		return "", rpc.ErrRpcUnauthenticated
	}

	return userID, nil
}
